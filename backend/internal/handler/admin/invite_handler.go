package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// InviteHandler handles admin invite management
type InviteHandler struct {
	inviteService         *service.InviteService
	adminActionLogService *service.AdminActionLogService
}

// NewInviteHandler creates a new admin invite handler
func NewInviteHandler(inviteService *service.InviteService, adminActionLogService *service.AdminActionLogService) *InviteHandler {
	return &InviteHandler{
		inviteService:         inviteService,
		adminActionLogService: adminActionLogService,
	}
}

type UpdateInviteSettingsRequest struct {
	RewardAmount float64 `json:"reward_amount" binding:"min=0"`
}

// GetSettings handles getting invite settings
// GET /api/v1/admin/invites/settings
func (h *InviteHandler) GetSettings(c *gin.Context) {
	settings, err := h.inviteService.GetInviteSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InviteSettings{RewardAmount: settings.RewardAmount})
}

// UpdateSettings handles updating invite settings
// PUT /api/v1/admin/invites/settings
func (h *InviteHandler) UpdateSettings(c *gin.Context) {
	var req UpdateInviteSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings, err := h.inviteService.UpdateInviteSettings(c.Request.Context(), service.InviteSettings{RewardAmount: req.RewardAmount})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InviteSettings{RewardAmount: settings.RewardAmount})
}

// ConfirmInvite handles confirming invite reward for an invitee
// POST /api/v1/admin/invites/:invitee_id/confirm
func (h *InviteHandler) ConfirmInvite(c *gin.Context) {
	inviteeID, err := strconv.ParseInt(c.Param("invitee_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invitee_id")
		return
	}

	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	invite, err := h.inviteService.ConfirmInvite(c.Request.Context(), inviteeID, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if invite != nil {
		payload := service.MarshalAdminActionPayload(map[string]any{
			"invite_id":     invite.ID,
			"inviter_id":    invite.InviterID,
			"invitee_id":    invite.InviteeID,
			"reward_amount": invite.RewardAmount,
		})
		inviteID := invite.ID
		h.adminActionLogService.Log(c.Request.Context(), service.AdminActionLogInput{
			AdminID:      &subject.UserID,
			Action:       "confirm_invite",
			ResourceType: "invite",
			ResourceID:   &inviteID,
			Payload:      payload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		})
	}

	response.Success(c, dto.InviteRecordFromService(invite))
}

// ListLogs handles listing invite logs
// GET /api/v1/admin/invites/logs
func (h *InviteHandler) ListLogs(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	filters := service.InviteLogFilters{
		Action:       strings.TrimSpace(c.Query("action")),
		InviterEmail: strings.TrimSpace(c.Query("inviter_email")),
		InviteeEmail: strings.TrimSpace(c.Query("invitee_email")),
	}

	if inviterIDStr := strings.TrimSpace(c.Query("inviter_id")); inviterIDStr != "" {
		inviterID, err := strconv.ParseInt(inviterIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid inviter_id")
			return
		}
		filters.InviterID = &inviterID
	}

	if inviteeIDStr := strings.TrimSpace(c.Query("invitee_id")); inviteeIDStr != "" {
		inviteeID, err := strconv.ParseInt(inviteeIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid invitee_id")
			return
		}
		filters.InviteeID = &inviteeID
	}

	if startStr := strings.TrimSpace(c.Query("start_time")); startStr != "" {
		startTime, err := parseInviteLogTime(startStr)
		if err != nil {
			response.BadRequest(c, "Invalid start_time")
			return
		}
		filters.StartTime = &startTime
	}

	if endStr := strings.TrimSpace(c.Query("end_time")); endStr != "" {
		endTime, err := parseInviteLogTime(endStr)
		if err != nil {
			response.BadRequest(c, "Invalid end_time")
			return
		}
		filters.EndTime = &endTime
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	logs, result, err := h.inviteService.ListInviteLogs(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]dto.InviteLog, 0, len(logs))
	for i := range logs {
		items = append(items, *dto.InviteLogFromService(&logs[i]))
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func parseInviteLogTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0).UTC(), nil
}
