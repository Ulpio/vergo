package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	s3store "github.com/Ulpio/vergo/internal/storage/s3"
)

type StorageHandler struct {
	s3 *s3store.S3
}

func NewStorageHandler(s3c *s3store.S3) *StorageHandler { return &StorageHandler{s3: s3c} }

type presignIn struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	ExpiresSec  int64  `json:"expires,omitempty"`
}

func (h *StorageHandler) Presign(c *gin.Context) {
	var in presignIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	url, headers, err := h.s3.PresignPut(c, in.Bucket, in.Key, in.ContentType, in.ExpiresSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "presign_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"method":  "PUT",
		"url":     url,
		"headers": headers,
	})
}
