package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/s3-web-manager/backend/internal/storage"
)

// ListBuckets handles GET /api/buckets
func ListBuckets(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		buckets, err := sp.ListAllBuckets(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to list buckets: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, buckets)
	}
}

// GetBucket handles GET /api/buckets/:name
func GetBucket(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		info, err := sp.GetBucketInfo(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, info)
	}
}
