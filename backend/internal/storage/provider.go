package storage

import (
	"context"
	"time"
)

// BucketInfo holds metadata about a bucket returned by admin-level operations.
type BucketInfo struct {
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	NumObjects int64  `json:"numObjects"`
	SizeBytes  int64  `json:"sizeBytes"`
}

// ObjectInfo holds metadata about a single object.
type ObjectInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

// ListResult is the response of ListObjects: directory-like prefixes and objects.
type ListResult struct {
	Prefixes []string     `json:"prefixes"`
	Objects  []ObjectInfo `json:"objects"`
}

// StorageProvider abstracts all S3-compatible storage operations.
// Implement this interface to add support for MinIO, AWS S3, etc.
type StorageProvider interface {
	// ListAllBuckets returns all buckets regardless of owner (admin-level).
	ListAllBuckets(ctx context.Context) ([]BucketInfo, error)
	// GetBucketInfo returns detailed info and stats for a single bucket.
	GetBucketInfo(ctx context.Context, bucket string) (BucketInfo, error)

	// ListObjects lists objects under the given prefix using delimiter "/" to simulate directories.
	ListObjects(ctx context.Context, bucket, prefix string) (*ListResult, error)
	// PresignUploadURL generates a pre-signed PUT URL for direct client-to-storage upload.
	PresignUploadURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error)
	// PresignDownloadURL generates a pre-signed GET URL for direct client-to-storage download.
	PresignDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	// DeleteObject deletes the specified object from a bucket.
	DeleteObject(ctx context.Context, bucket, key string) error
}
