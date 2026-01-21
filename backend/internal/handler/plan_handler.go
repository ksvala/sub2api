package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlanHandler handles public plan listing
type PlanHandler struct {
	planService *service.PlanService
}

// NewPlanHandler creates a new plan handler
func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// List handles listing enabled plans
// GET /api/v1/plans
func (h *PlanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	enabled := true

	plans, result, err := h.planService.List(c.Request.Context(), params, &enabled)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]dto.PublicPlan, 0, len(plans))
	for i := range plans {
		items = append(items, *dto.PublicPlanFromService(&plans[i]))
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}
