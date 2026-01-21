package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles admin file uploads
type UploadHandler struct {
	uploadService        *service.UploadService
	adminActionLogService *service.AdminActionLogService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadService *service.UploadService, adminActionLogService *service.AdminActionLogService) *UploadHandler {
	return &UploadHandler{
		uploadService:        uploadService,
		adminActionLogService: adminActionLogService,
	}
}

// UploadImage handles image upload
// POST /api/v1/admin/uploads/image
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	url, err := h.uploadService.SaveImage(c.Request.Context(), file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		payload := service.MarshalAdminActionPayload(map[string]any{
			"url": url,
		})
		h.adminActionLogService.Log(c.Request.Context(), service.AdminActionLogInput{
			AdminID:      &subject.UserID,
			Action:       "upload_image",
			ResourceType: "upload",
			Payload:      payload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		})
	}

	response.Success(c, gin.H{"url": url})
}
