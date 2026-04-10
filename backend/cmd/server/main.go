package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/s3-web-manager/backend/internal/api"
	"github.com/s3-web-manager/backend/internal/config"
	"github.com/s3-web-manager/backend/internal/middleware"
	"github.com/s3-web-manager/backend/internal/storage"
	"github.com/s3-web-manager/backend/internal/storage/ceph"
)

func main() {
	// Load .env if present (optional; Docker passes env vars directly)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	sp, err := newStorageProvider(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage provider: %v", err)
	}

	r := setupRouter(cfg, sp)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("starting s3-web-manager backend on %s (storage: %s)", addr, cfg.StorageType)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// newStorageProvider selects and initializes the storage provider based on STORAGE_TYPE.
func newStorageProvider(cfg *config.Config) (storage.StorageProvider, error) {
	switch cfg.StorageType {
	case "ceph":
		return ceph.NewCephProvider(cfg.CephEndpoint, cfg.CephAccessKey, cfg.CephSecretKey, cfg.CephRegion)
	default:
		return nil, fmt.Errorf("unsupported STORAGE_TYPE %q (supported: ceph)", cfg.StorageType)
	}
}

// setupRouter builds the Gin router with all routes registered.
func setupRouter(cfg *config.Config, sp storage.StorageProvider) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Public routes
	r.POST("/api/auth/login", api.Login(cfg))

	// Protected routes
	protected := r.Group("/api", middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.GET("/buckets", api.ListBuckets(sp))
		protected.GET("/buckets/:name", api.GetBucket(sp))
		protected.GET("/buckets/:name/objects", api.ListObjects(sp))
		protected.POST("/buckets/:name/objects/upload-url", api.GetUploadURL(sp))
		protected.GET("/buckets/:name/objects/download-url", api.GetDownloadURL(sp))
		protected.DELETE("/buckets/:name/objects", api.DeleteObject(sp))
	}

	return r
}
