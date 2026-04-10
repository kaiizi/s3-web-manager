package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/s3-web-manager/backend/internal/api"
	"github.com/s3-web-manager/backend/internal/storage"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockProvider is a test implementation of storage.StorageProvider.
type mockProvider struct {
	listAllBuckets     func(ctx context.Context) ([]storage.BucketInfo, error)
	getBucketInfo      func(ctx context.Context, bucket string) (storage.BucketInfo, error)
	listObjects        func(ctx context.Context, bucket, prefix string) (*storage.ListResult, error)
	presignUploadURL   func(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error)
	presignDownloadURL func(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	deleteObject       func(ctx context.Context, bucket, key string) error
}

func (m *mockProvider) ListAllBuckets(ctx context.Context) ([]storage.BucketInfo, error) {
	return m.listAllBuckets(ctx)
}
func (m *mockProvider) GetBucketInfo(ctx context.Context, bucket string) (storage.BucketInfo, error) {
	return m.getBucketInfo(ctx, bucket)
}
func (m *mockProvider) ListObjects(ctx context.Context, bucket, prefix string) (*storage.ListResult, error) {
	return m.listObjects(ctx, bucket, prefix)
}
func (m *mockProvider) PresignUploadURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error) {
	return m.presignUploadURL(ctx, bucket, key, contentType, expiry)
}
func (m *mockProvider) PresignDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return m.presignDownloadURL(ctx, bucket, key, expiry)
}
func (m *mockProvider) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.deleteObject(ctx, bucket, key)
}

// ── Bucket handlers ──────────────────────────────────────────────────────────

func TestListBuckets_Success(t *testing.T) {
	mp := &mockProvider{
		listAllBuckets: func(ctx context.Context) ([]storage.BucketInfo, error) {
			return []storage.BucketInfo{
				{Name: "bucket-a", Owner: "alice", NumObjects: 3, SizeBytes: 1024},
			}, nil
		},
	}

	r := gin.New()
	r.GET("/api/buckets", api.ListBuckets(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var result []storage.BucketInfo
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 1 || result[0].Name != "bucket-a" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestListBuckets_StorageError(t *testing.T) {
	mp := &mockProvider{
		listAllBuckets: func(ctx context.Context) ([]storage.BucketInfo, error) {
			return nil, errors.New("ceph down")
		},
	}

	r := gin.New()
	r.GET("/api/buckets", api.ListBuckets(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status: got %d, want 502", w.Code)
	}
}

func TestGetBucket_Success(t *testing.T) {
	mp := &mockProvider{
		getBucketInfo: func(ctx context.Context, bucket string) (storage.BucketInfo, error) {
			return storage.BucketInfo{Name: bucket, Owner: "bob"}, nil
		},
	}

	r := gin.New()
	r.GET("/api/buckets/:name", api.GetBucket(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/my-bucket", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}

func TestGetBucket_NotFound(t *testing.T) {
	mp := &mockProvider{
		getBucketInfo: func(ctx context.Context, bucket string) (storage.BucketInfo, error) {
			return storage.BucketInfo{}, errors.New("not found")
		},
	}

	r := gin.New()
	r.GET("/api/buckets/:name", api.GetBucket(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/ghost", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}

// ── Object handlers ───────────────────────────────────────────────────────────

func TestListObjects_Success(t *testing.T) {
	mp := &mockProvider{
		listObjects: func(ctx context.Context, bucket, prefix string) (*storage.ListResult, error) {
			return &storage.ListResult{
				Prefixes: []string{"folder/"},
				Objects: []storage.ObjectInfo{
					{Key: "file.txt", Size: 100},
				},
			}, nil
		},
	}

	r := gin.New()
	r.GET("/api/buckets/:name/objects", api.ListObjects(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test/objects?prefix=", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var result storage.ListResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Prefixes) != 1 || result.Prefixes[0] != "folder/" {
		t.Errorf("unexpected prefixes: %v", result.Prefixes)
	}
	if len(result.Objects) != 1 || result.Objects[0].Key != "file.txt" {
		t.Errorf("unexpected objects: %v", result.Objects)
	}
}

func TestListObjects_Error(t *testing.T) {
	mp := &mockProvider{
		listObjects: func(ctx context.Context, bucket, prefix string) (*storage.ListResult, error) {
			return nil, errors.New("access denied")
		},
	}

	r := gin.New()
	r.GET("/api/buckets/:name/objects", api.ListObjects(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test/objects?prefix=", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want 500", w.Code)
	}
}

func TestGetUploadURL_Success(t *testing.T) {
	mp := &mockProvider{
		presignUploadURL: func(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error) {
			return "http://ceph/presigned-put", nil
		},
	}

	r := gin.New()
	r.POST("/api/buckets/:name/objects/upload-url", api.GetUploadURL(mp))

	body := strings.NewReader(`{"key":"test.txt","contentType":"text/plain"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/buckets/test/objects/upload-url", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["url"] != "http://ceph/presigned-put" {
		t.Errorf("unexpected url: %q", result["url"])
	}
	if result["method"] != "PUT" {
		t.Errorf("unexpected method: %q", result["method"])
	}
}

func TestGetUploadURL_MissingBody(t *testing.T) {
	mp := &mockProvider{}

	r := gin.New()
	r.POST("/api/buckets/:name/objects/upload-url", api.GetUploadURL(mp))

	req := httptest.NewRequest(http.MethodPost, "/api/buckets/test/objects/upload-url", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}

func TestGetDownloadURL_Success(t *testing.T) {
	mp := &mockProvider{
		presignDownloadURL: func(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
			return "http://ceph/presigned-get", nil
		},
	}

	r := gin.New()
	r.GET("/api/buckets/:name/objects/download-url", api.GetDownloadURL(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test/objects/download-url?key=file.txt", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}

func TestGetDownloadURL_MissingKey(t *testing.T) {
	mp := &mockProvider{}

	r := gin.New()
	r.GET("/api/buckets/:name/objects/download-url", api.GetDownloadURL(mp))

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test/objects/download-url", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}

func TestDeleteObject_Success(t *testing.T) {
	deleted := ""
	mp := &mockProvider{
		deleteObject: func(ctx context.Context, bucket, key string) error {
			deleted = bucket + "/" + key
			return nil
		},
	}

	r := gin.New()
	r.DELETE("/api/buckets/:name/objects", api.DeleteObject(mp))

	req := httptest.NewRequest(http.MethodDelete, "/api/buckets/test/objects?key=file.txt", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", w.Code)
	}
	if deleted != "test/file.txt" {
		t.Errorf("deleted: got %q, want %q", deleted, "test/file.txt")
	}
}

func TestDeleteObject_MissingKey(t *testing.T) {
	mp := &mockProvider{}

	r := gin.New()
	r.DELETE("/api/buckets/:name/objects", api.DeleteObject(mp))

	req := httptest.NewRequest(http.MethodDelete, "/api/buckets/test/objects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}
