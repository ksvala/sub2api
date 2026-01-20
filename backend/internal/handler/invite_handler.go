package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteHandler handles invite-related user requests
type InviteHandler struct {
	inviteService *service.InviteService
}

// NewInviteHandler creates a new InviteHandler
func NewInviteHandler(inviteService *service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
	}
}

// GetSummary handles getting invite summary for current user
// GET /api/v1/invites/summary
func (h *InviteHandler) GetSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	summary, err := h.inviteService.GetInviteSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InviteSummaryFromService(summary))
}

// ListRecords handles listing invite records for current user
// GET /api/v1/invites/records
func (h *InviteHandler) ListRecords(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	invites, result, err := h.inviteService.ListInvitesByInviter(c.Request.Context(), subject.UserID, params, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]dto.InviteRecord, 0, len(invites))
	for i := range invites {
		items = append(items, *dto.InviteRecordFromService(&invites[i]))
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

// ListRewards handles listing invite reward records for current user
// GET /api/v1/invites/rewards
func (h *InviteHandler) ListRewards(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	rewards, result, err := h.inviteService.ListInviteRewardRecords(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]dto.InviteRewardRecord, 0, len(rewards))
	for i := range rewards {
		items = append(items, *dto.InviteRewardRecordFromRedeem(&rewards[i]))
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}
