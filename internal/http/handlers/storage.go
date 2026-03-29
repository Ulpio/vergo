package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/domain/file"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/pkg/config"
	s3store "github.com/Ulpio/vergo/internal/storage/s3"
)

type StorageHandler struct {
	s3 *s3store.S3
	fs file.Service
}

func NewStorageHandler(s3c *s3store.S3, fs file.Service) *StorageHandler {
	return &StorageHandler{s3: s3c, fs: fs}
}

// ---------- Presign (PUT) ----------

type presignPutIn struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	ExpiresSec  int64  `json:"expires,omitempty"`
}

// PresignPut generates a presigned URL for S3 upload.
// @Summary Get presigned upload URL
// @Tags Storage
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body presignPutIn true "Upload parameters"
// @Success 200 {object} PresignPutResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorDetailResponse
// @Router /storage/presign [post]
func (h *StorageHandler) PresignPut(c *gin.Context) {
	var in presignPutIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	url, headers, err := h.s3.PresignPut(c.Request.Context(), in.Bucket, in.Key, in.ContentType, in.ExpiresSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "presign_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"method": "PUT", "url": url, "headers": headers})
}

// ---------- Presign (GET) ----------

type presignGetIn struct {
	Bucket     string `json:"bucket"`
	Key        string `json:"key" binding:"required"`
	ExpiresSec int64  `json:"expires,omitempty"`
}

// PresignGet generates a presigned URL for S3 download.
// @Summary Get presigned download URL
// @Tags Storage
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body presignGetIn true "Download parameters"
// @Success 200 {object} PresignGetResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorDetailResponse
// @Router /storage/presign-download [post]
func (h *StorageHandler) PresignGet(c *gin.Context) {
	var in presignGetIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	url, err := h.s3.PresignGet(c.Request.Context(), in.Bucket, in.Key, in.ExpiresSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "presign_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"method": "GET", "url": url})
}

type fileCreateIn struct {
	Bucket      string      `json:"bucket,omitempty"`
	Key         string      `json:"key" binding:"required"`
	SizeBytes   *int64      `json:"size_bytes,omitempty"`
	ContentType string      `json:"content_type,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// CreateFile registers file metadata after upload.
// @Summary Register uploaded file
// @Tags Storage
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body fileCreateIn true "File metadata"
// @Success 201 {object} file.File
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /storage/files [post]
func (h *StorageHandler) CreateFile(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}

	var in fileCreateIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	cfg := config.Load()
	if len(cfg.StorageAllowedTypes) > 0 && in.ContentType != "" {
		okType := false
		for _, t := range cfg.StorageAllowedTypes {
			if t == in.ContentType {
				okType = true
				break
			}
		}
		if !okType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_content_type"})
			return
		}
	}
	if in.SizeBytes != nil && cfg.StorageMaxMB > 0 {
		maxBytes := int64(cfg.StorageMaxMB) * 1024 * 1024
		if *in.SizeBytes > maxBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large", "max_mb": cfg.StorageMaxMB})
			return
		}
	}

	key := in.Key
	if !strings.HasPrefix(key, "org/") {
		key = fmt.Sprintf("org/%s/users/%s/%s", orgID, uid, strings.TrimLeft(in.Key, "/"))
	}

	f, err := h.fs.Create(orgID, uid, in.Bucket, key, in.SizeBytes, in.ContentType, in.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
		return
	}
	c.JSON(http.StatusCreated, f)
}

// ListFiles returns files in the organization.
// @Summary List files
// @Tags Storage
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param limit query int false "Items per page (max 100)" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "items, next_offset"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /storage/files [get]
func (h *StorageHandler) ListFiles(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.fs.List(file.ListParams{OrgID: orgID, Limit: limit, Offset: offset})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "next_offset": offset + len(items)})
}

// GetFile returns file metadata by ID.
// @Summary Get file
// @Tags Storage
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "File ID"
// @Success 200 {object} file.File
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /storage/files/{id} [get]
func (h *StorageHandler) GetFile(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	id := c.Param("id")

	f, err := h.fs.Get(orgID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, f)
}

// DeleteFile removes a file from S3 and database.
// @Summary Delete file
// @Tags Storage
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "File ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /storage/files/{id} [delete]
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	id := c.Param("id")

	// Primeiro busca metadados para saber bucket/key
	f, err := h.fs.Get(orgID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	// Deleta no S3 (best-effort)
	if err := h.s3.DeleteObject(c.Request.Context(), f.Bucket, f.ObjectKey); err != nil {
		slog.ErrorContext(c.Request.Context(), "s3 delete failed", "error", err)
	}

	// Remove metadados
	if err := h.fs.Delete(orgID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed"})
		return
	}
	c.Status(http.StatusNoContent)
}
