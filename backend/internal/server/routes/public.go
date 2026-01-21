package routes

import "github.com/Wei-Shaw/sub2api/internal/handler"

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes registers public API v1 routes that do not require auth.
func RegisterPublicRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	// Plans (public): used by marketing /home pricing section.
	v1.GET("/plans", h.Plan.List)
}
