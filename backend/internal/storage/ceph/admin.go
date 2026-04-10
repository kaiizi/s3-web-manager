package ceph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	signerv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/s3-web-manager/backend/internal/storage"
)

// adminClient makes signed HTTP requests to the Ceph RGW Admin OPS API.
type adminClient struct {
	endpoint   string
	accessKey  string
	secretKey  string
	region     string
	httpClient *http.Client
}

func newAdminClient(endpoint, accessKey, secretKey, region string) *adminClient {
	return &adminClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		accessKey:  accessKey,
		secretKey:  secretKey,
		region:     region,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// adminBucketStats is the JSON response from GET /admin/bucket?bucket=<name>&stats=true
type adminBucketStats struct {
	Bucket string `json:"bucket"`
	Owner  string `json:"owner"`
	Usage  struct {
		RGWMain struct {
			NumObjects int64 `json:"num_objects"`
			SizeActual int64 `json:"size_actual"`
		} `json:"rgw.main"`
	} `json:"usage"`
}

// jsonUserInfo is the JSON response from GET /admin/user?uid=<uid>
type jsonUserInfo struct {
	Keys []struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
	} `json:"keys"`
}

// UserKeys holds the S3 credentials for a Ceph user.
type UserKeys struct {
	AccessKey string
	SecretKey string
}

// GetUserKeys fetches the S3 access/secret key for the given Ceph user uid.
func (c *adminClient) GetUserKeys(ctx context.Context, uid string) (*UserKeys, error) {
	body, err := c.doRequest(ctx, "GET", "/admin/user", url.Values{"uid": {uid}})
	if err != nil {
		return nil, fmt.Errorf("admin get user %q: %w", uid, err)
	}

	var info jsonUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse user info for %q: %w (body: %.200s)", uid, err, string(body))
	}
	if len(info.Keys) == 0 {
		return nil, fmt.Errorf("user %q has no S3 keys", uid)
	}
	return &UserKeys{
		AccessKey: info.Keys[0].AccessKey,
		SecretKey: info.Keys[0].SecretKey,
	}, nil
}

// ListAllBuckets implements StorageProvider via Admin OPS API.
func (p *CephProvider) ListAllBuckets(ctx context.Context) ([]storage.BucketInfo, error) {
	// GET /admin/bucket?list returns XML: <buckets><bucket>name</bucket>...</buckets>
	body, err := p.adminClient.doRequest(ctx, "GET", "/admin/bucket", url.Values{"list": {""}})
	if err != nil {
		return nil, err
	}

	var bucketNames []string
	if err := json.Unmarshal(body, &bucketNames); err != nil {
		return nil, fmt.Errorf("failed to parse bucket list: %w (body: %.200s)", err, string(body))
	}

	result := make([]storage.BucketInfo, 0, len(bucketNames))
	for _, name := range bucketNames {
		info, err := p.GetBucketInfo(ctx, name)
		if err != nil {
			// Include bucket with partial info rather than failing the whole list
			result = append(result, storage.BucketInfo{Name: name})
			continue
		}
		result = append(result, info)
	}
	return result, nil
}

// GetBucketInfo implements StorageProvider via Admin OPS API stats.
func (p *CephProvider) GetBucketInfo(ctx context.Context, bucket string) (storage.BucketInfo, error) {
	params := url.Values{
		"bucket": {bucket},
		"stats":  {"true"},
	}
	body, err := p.adminClient.doRequest(ctx, "GET", "/admin/bucket", params)
	if err != nil {
		return storage.BucketInfo{}, err
	}

	var stats adminBucketStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return storage.BucketInfo{}, fmt.Errorf("failed to parse bucket stats: %w (body: %.200s)", err, string(body))
	}

	return storage.BucketInfo{
		Name:       stats.Bucket,
		Owner:      stats.Owner,
		NumObjects: stats.Usage.RGWMain.NumObjects,
		SizeBytes:  stats.Usage.RGWMain.SizeActual,
	}, nil
}

// doRequest performs an AWS v4 signed request to the Admin OPS API.
func (c *adminClient) doRequest(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	rawQuery := params.Encode()
	fullURL := fmt.Sprintf("%s%s?%s", c.endpoint, path, rawQuery)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Ceph RGW requires x-amz-content-sha256 to be present and signed
	emptyBodyHash := sha256Hex("")
	req.Header.Set("x-amz-content-sha256", emptyBodyHash)

	// Sign using the official aws-sdk-go-v2 signer to ensure correctness
	creds := aws.Credentials{
		AccessKeyID:     c.accessKey,
		SecretAccessKey: c.secretKey,
	}
	signer := signerv4.NewSigner()
	if err := signer.SignHTTP(ctx, creds, req, emptyBodyHash, "s3", c.region, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("admin API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
