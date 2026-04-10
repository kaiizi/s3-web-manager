package ceph

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/s3-web-manager/backend/internal/storage"
)

// CephProvider implements StorageProvider for Ceph RGW v18.4.
// Standard S3 operations use aws-sdk-go-v2; admin operations use the Ceph Admin OPS API.
type CephProvider struct {
	adminClient *adminClient
	endpoint    string
	region      string

	// ownerClientCache caches per-user S3 clients keyed by owner uid.
	// This avoids fetching keys and re-creating clients on every request.
	ownerClientMu    sync.Mutex
	ownerClientCache map[string]*s3.Client
}

// NewCephProvider creates a CephProvider configured for Ceph RGW.
func NewCephProvider(endpoint, accessKey, secretKey, region string) (*CephProvider, error) {
	return &CephProvider{
		adminClient:      newAdminClient(endpoint, accessKey, secretKey, region),
		endpoint:         endpoint,
		region:           region,
		ownerClientCache: make(map[string]*s3.Client),
	}, nil
}

// newS3Client creates a path-style S3 client for the given credentials.
func (p *CephProvider) newS3Client(accessKey, secretKey string) (*s3.Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               p.endpoint,
				SigningRegion:     p.region,
				HostnameImmutable: true,
			}, nil
		},
	)
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(p.region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

// s3ClientForBucket returns an S3 client authenticated as the bucket owner.
// Owner keys are fetched via Admin OPS API and cached per-uid.
func (p *CephProvider) s3ClientForBucket(ctx context.Context, bucket string) (*s3.Client, error) {
	// Get bucket owner
	info, err := p.GetBucketInfo(ctx, bucket)
	if err != nil {
		return nil, err
	}
	owner := info.Owner

	// Check cache first (fast path without lock)
	p.ownerClientMu.Lock()
	defer p.ownerClientMu.Unlock()
	if c, ok := p.ownerClientCache[owner]; ok {
		return c, nil
	}

	// Fetch owner's S3 keys via Admin API
	keys, err := p.adminClient.GetUserKeys(ctx, owner)
	if err != nil {
		return nil, err
	}
	c, err := p.newS3Client(keys.AccessKey, keys.SecretKey)
	if err != nil {
		return nil, err
	}
	p.ownerClientCache[owner] = c
	return c, nil
}

// ListObjects implements StorageProvider.
func (p *CephProvider) ListObjects(ctx context.Context, bucket, prefix string) (*storage.ListResult, error) {
	sc, err := p.s3ClientForBucket(ctx, bucket)
	if err != nil {
		return nil, err
	}

	out, err := sc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	result := &storage.ListResult{
		Prefixes: make([]string, 0, len(out.CommonPrefixes)),
		Objects:  make([]storage.ObjectInfo, 0, len(out.Contents)),
	}
	for _, cp := range out.CommonPrefixes {
		if cp.Prefix != nil {
			result.Prefixes = append(result.Prefixes, *cp.Prefix)
		}
	}
	for _, obj := range out.Contents {
		var info storage.ObjectInfo
		if obj.Key != nil {
			info.Key = *obj.Key
		}
		if obj.Size != nil {
			info.Size = *obj.Size
		}
		if obj.LastModified != nil {
			info.LastModified = *obj.LastModified
		}
		result.Objects = append(result.Objects, info)
	}
	return result, nil
}

// PresignUploadURL implements StorageProvider.
func (p *CephProvider) PresignUploadURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error) {
	sc, err := p.s3ClientForBucket(ctx, bucket)
	if err != nil {
		return "", err
	}
	req, err := s3.NewPresignClient(sc).PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// PresignDownloadURL implements StorageProvider.
func (p *CephProvider) PresignDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	sc, err := p.s3ClientForBucket(ctx, bucket)
	if err != nil {
		return "", err
	}
	req, err := s3.NewPresignClient(sc).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// DeleteObject implements StorageProvider.
func (p *CephProvider) DeleteObject(ctx context.Context, bucket, key string) error {
	sc, err := p.s3ClientForBucket(ctx, bucket)
	if err != nil {
		return err
	}
	_, err = sc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}
