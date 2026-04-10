package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/s3-web-manager/backend/internal/storage"
)

const presignExpiry = 15 * time.Minute

// ListObjects handles GET /api/buckets/:name/objects?prefix=<prefix>
func ListObjects(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := c.Param("name")
		prefix := c.DefaultQuery("prefix", "")

		result, err := sp.ListObjects(c.Request.Context(), bucket, prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list objects: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

type uploadURLRequest struct {
	Key         string `json:"key" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// GetUploadURL handles POST /api/buckets/:name/objects/upload-url
func GetUploadURL(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := c.Param("name")
		var req uploadURLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key and contentType are required"})
			return
		}

		presignedURL, err := sp.PresignUploadURL(c.Request.Context(), bucket, req.Key, req.ContentType, presignExpiry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate upload URL: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": presignedURL, "method": "PUT"})
	}
}

// GetDownloadURL handles GET /api/buckets/:name/objects/download-url?key=<key>
func GetDownloadURL(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := c.Param("name")
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
			return
		}

		presignedURL, err := sp.PresignDownloadURL(c.Request.Context(), bucket, key, presignExpiry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate download URL: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": presignedURL})
	}
}

// DeleteObject handles DELETE /api/buckets/:name/objects?key=<key>
func DeleteObject(sp storage.StorageProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := c.Param("name")
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
			return
		}

		if err := sp.DeleteObject(c.Request.Context(), bucket, key); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete object: " + err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
