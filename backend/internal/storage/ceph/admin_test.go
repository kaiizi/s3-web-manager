package ceph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// newTestAdminClient creates an adminClient pointed at the given test server URL.
func newTestAdminClient(url string) *adminClient {
	return newAdminClient(url, "test-ak", "test-sk", "default")
}

func TestGetUserKeys_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/user" || r.URL.Query().Get("uid") != "alice" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"user_id":"alice","keys":[{"user":"alice","access_key":"AKEYALICE","secret_key":"SKEYALICE"}]}`))
	}))
	defer srv.Close()

	ac := newTestAdminClient(srv.URL)
	keys, err := ac.GetUserKeys(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys.AccessKey != "AKEYALICE" {
		t.Errorf("access key: got %q, want %q", keys.AccessKey, "AKEYALICE")
	}
	if keys.SecretKey != "SKEYALICE" {
		t.Errorf("secret key: got %q, want %q", keys.SecretKey, "SKEYALICE")
	}
}

func TestGetUserKeys_NoKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"user_id":"ghost","keys":[]}`))
	}))
	defer srv.Close()

	ac := newTestAdminClient(srv.URL)
	_, err := ac.GetUserKeys(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for user with no keys, got nil")
	}
}

func TestGetUserKeys_AdminError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"Code":"NoSuchUser"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	ac := newTestAdminClient(srv.URL)
	_, err := ac.GetUserKeys(context.Background(), "nobody")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestListAllBuckets_ParsesJSON(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if _, hasListParam := r.URL.Query()["list"]; hasListParam {
			// First call: return bucket list
			w.Write([]byte(`["bucket-a","bucket-b"]`))
			return
		}
		// Per-bucket stats call
		bucket := r.URL.Query().Get("bucket")
		w.Write([]byte(`{"bucket":"` + bucket + `","owner":"owner-` + bucket + `","usage":{"rgw.main":{"num_objects":5,"size_actual":1024}}}`))
	}))
	defer srv.Close()

	p := &CephProvider{
		adminClient:      newTestAdminClient(srv.URL),
		endpoint:         srv.URL,
		region:           "default",
		ownerClientCache: make(map[string]*s3.Client),
	}
	buckets, err := p.ListAllBuckets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if buckets[0].Name != "bucket-a" {
		t.Errorf("bucket[0].Name: got %q, want %q", buckets[0].Name, "bucket-a")
	}
	if buckets[0].Owner != "owner-bucket-a" {
		t.Errorf("bucket[0].Owner: got %q, want %q", buckets[0].Owner, "owner-bucket-a")
	}
	if buckets[0].NumObjects != 5 {
		t.Errorf("bucket[0].NumObjects: got %d, want 5", buckets[0].NumObjects)
	}
}

func TestGetBucketInfo_ParsesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"bucket":"mybucket","owner":"myowner","usage":{"rgw.main":{"num_objects":42,"size_actual":8192}}}`))
	}))
	defer srv.Close()

	p := &CephProvider{
		adminClient:      newTestAdminClient(srv.URL),
		endpoint:         srv.URL,
		region:           "default",
		ownerClientCache: make(map[string]*s3.Client),
	}
	info, err := p.GetBucketInfo(context.Background(), "mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "mybucket" {
		t.Errorf("Name: got %q, want %q", info.Name, "mybucket")
	}
	if info.Owner != "myowner" {
		t.Errorf("Owner: got %q, want %q", info.Owner, "myowner")
	}
	if info.NumObjects != 42 {
		t.Errorf("NumObjects: got %d, want 42", info.NumObjects)
	}
	if info.SizeBytes != 8192 {
		t.Errorf("SizeBytes: got %d, want 8192", info.SizeBytes)
	}
}
