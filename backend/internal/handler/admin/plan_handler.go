package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlanHandler handles admin plan management
type PlanHandler struct {
	planService           *service.PlanService
	adminActionLogService *service.AdminActionLogService
}

// NewPlanHandler creates a new admin plan handler
func NewPlanHandler(planService *service.PlanService, adminActionLogService *service.AdminActionLogService) *PlanHandler {
	return &PlanHandler{
		planService:           planService,
		adminActionLogService: adminActionLogService,
	}
}

type CreatePlanRequest struct {
	Title         string  `json:"title" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,min=0"`
	GroupName     string  `json:"group_name"`
	GroupSort     int     `json:"group_sort"`
	DailyQuota    float64 `json:"daily_quota"`
	TotalQuota    float64 `json:"total_quota"`
	PurchaseQRURL string  `json:"purchase_qr_url"`
	Enabled       *bool   `json:"enabled"`
	SortOrder     int     `json:"sort_order"`
}

type UpdatePlanRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	Price         *float64 `json:"price" binding:"omitempty,min=0"`
	GroupName     *string  `json:"group_name"`
	GroupSort     *int     `json:"group_sort"`
	DailyQuota    *float64 `json:"daily_quota" binding:"omitempty,min=0"`
	TotalQuota    *float64 `json:"total_quota" binding:"omitempty,min=0"`
	PurchaseQRURL *string  `json:"purchase_qr_url"`
	Enabled       *bool    `json:"enabled"`
	SortOrder     *int     `json:"sort_order"`
}

// List handles listing plans (admin)
// GET /api/v1/admin/plans
func (h *PlanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	var enabled *bool
	if raw := strings.TrimSpace(c.Query("enabled")); raw != "" {
		val := raw == "true" || raw == "1"
		enabled = &val
	}

	plans, result, err := h.planService.List(c.Request.Context(), params, enabled)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]dto.Plan, 0, len(plans))
	for i := range plans {
		items = append(items, *dto.PlanFromService(&plans[i]))
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

// GetByID handles getting a plan by ID
// GET /api/v1/admin/plans/:id
func (h *PlanHandler) GetByID(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.planService.GetByID(c.Request.Context(), planID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PlanFromService(plan))
}

// Create handles creating a plan
// POST /api/v1/admin/plans
func (h *PlanHandler) Create(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	plan, err := h.planService.Create(c.Request.Context(), &service.CreatePlanInput{
		Title:         req.Title,
		Description:   req.Description,
		Price:         req.Price,
		GroupName:     req.GroupName,
		GroupSort:     req.GroupSort,
		DailyQuota:    req.DailyQuota,
		TotalQuota:    req.TotalQuota,
		PurchaseQRURL: req.PurchaseQRURL,
		Enabled:       enabled,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		payload := service.MarshalAdminActionPayload(map[string]any{
			"title":        plan.Title,
			"price":        plan.Price,
			"group_name":   plan.GroupName,
			"group_sort":   plan.GroupSort,
			"daily_quota":  plan.DailyQuota,
			"total_quota":  plan.TotalQuota,
			"enabled":      plan.Enabled,
			"sort_order":   plan.SortOrder,
			"purchase_qr":  plan.PurchaseQRURL,
		})
		planID := plan.ID
		h.adminActionLogService.Log(c.Request.Context(), service.AdminActionLogInput{
			AdminID:      &subject.UserID,
			Action:       "create_plan",
			ResourceType: "plan",
			ResourceID:   &planID,
			Payload:      payload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		})
	}

	response.Success(c, dto.PlanFromService(plan))
}

// Update handles updating a plan
// PUT /api/v1/admin/plans/:id
func (h *PlanHandler) Update(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.Update(c.Request.Context(), planID, &service.UpdatePlanInput{
		Title:         req.Title,
		Description:   req.Description,
		Price:         req.Price,
		GroupName:     req.GroupName,
		GroupSort:     req.GroupSort,
		DailyQuota:    req.DailyQuota,
		TotalQuota:    req.TotalQuota,
		PurchaseQRURL: req.PurchaseQRURL,
		Enabled:       req.Enabled,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		payload := service.MarshalAdminActionPayload(map[string]any{
			"title":        plan.Title,
			"price":        plan.Price,
			"group_name":   plan.GroupName,
			"group_sort":   plan.GroupSort,
			"daily_quota":  plan.DailyQuota,
			"total_quota":  plan.TotalQuota,
			"enabled":      plan.Enabled,
			"sort_order":   plan.SortOrder,
			"purchase_qr":  plan.PurchaseQRURL,
		})
		planID := plan.ID
		h.adminActionLogService.Log(c.Request.Context(), service.AdminActionLogInput{
			AdminID:      &subject.UserID,
			Action:       "update_plan",
			ResourceType: "plan",
			ResourceID:   &planID,
			Payload:      payload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		})
	}

	response.Success(c, dto.PlanFromService(plan))
}

// Delete handles deleting a plan
// DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) Delete(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.planService.Delete(c.Request.Context(), planID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		payload := service.MarshalAdminActionPayload(map[string]any{
			"plan_id": planID,
		})
		h.adminActionLogService.Log(c.Request.Context(), service.AdminActionLogInput{
			AdminID:      &subject.UserID,
			Action:       "delete_plan",
			ResourceType: "plan",
			ResourceID:   &planID,
			Payload:      payload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		})
	}

	response.Success(c, gin.H{"message": "Plan deleted successfully"})
}
