package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type InviteSummary struct {
	InviteCode        string  `json:"invite_code"`
	TotalInvites      int     `json:"total_invites"`
	PendingInvites    int     `json:"pending_invites"`
	ConfirmedInvites  int     `json:"confirmed_invites"`
	TotalRewardAmount float64 `json:"total_reward_amount"`
}

type InviteRecord struct {
	ID           int64      `json:"id"`
	InviteeEmail string     `json:"invitee_email"`
	CreatedAt    time.Time  `json:"created_at"`
	Status       string     `json:"status"`
	ConfirmedAt  *time.Time `json:"confirmed_at,omitempty"`
	RewardAmount float64    `json:"reward_amount"`
}

type InviteRewardRecord struct {
	ID        int64     `json:"id"`
	Amount    float64   `json:"amount"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type InviteLog struct {
	ID           int64     `json:"id"`
	Action       string    `json:"action"`
	InviterID    int64     `json:"inviter_id"`
	InviterEmail string    `json:"inviter_email"`
	InviteeID    int64     `json:"invitee_id"`
	InviteeEmail string    `json:"invitee_email"`
	AdminID      *int64    `json:"admin_id,omitempty"`
	AdminEmail   string    `json:"admin_email,omitempty"`
	RewardAmount float64   `json:"reward_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

type InviteSettings struct {
	RewardAmount float64 `json:"reward_amount"`
}

func InviteSummaryFromService(summary *service.InviteSummary) *InviteSummary {
	if summary == nil {
		return nil
	}
	return &InviteSummary{
		InviteCode:        summary.InviteCode,
		TotalInvites:      summary.TotalInvites,
		PendingInvites:    summary.PendingInvites,
		ConfirmedInvites:  summary.ConfirmedInvites,
		TotalRewardAmount: summary.TotalRewardAmount,
	}
}

func InviteRecordFromService(invite *service.Invite) *InviteRecord {
	if invite == nil {
		return nil
	}
	inviteeEmail := ""
	if invite.Invitee != nil {
		inviteeEmail = invite.Invitee.Email
	}
	return &InviteRecord{
		ID:           invite.ID,
		InviteeEmail: inviteeEmail,
		CreatedAt:    invite.CreatedAt,
		Status:       invite.Status,
		ConfirmedAt:  invite.ConfirmedAt,
		RewardAmount: invite.RewardAmount,
	}
}

func InviteLogFromService(logRecord *service.InviteLog) *InviteLog {
	if logRecord == nil {
		return nil
	}
	inviterEmail := ""
	inviteeEmail := ""
	adminEmail := ""
	if logRecord.Inviter != nil {
		inviterEmail = logRecord.Inviter.Email
	}
	if logRecord.Invitee != nil {
		inviteeEmail = logRecord.Invitee.Email
	}
	if logRecord.Admin != nil {
		adminEmail = logRecord.Admin.Email
	}
	return &InviteLog{
		ID:           logRecord.ID,
		Action:       logRecord.Action,
		InviterID:    logRecord.InviterID,
		InviterEmail: inviterEmail,
		InviteeID:    logRecord.InviteeID,
		InviteeEmail: inviteeEmail,
		AdminID:      logRecord.AdminID,
		AdminEmail:   adminEmail,
		RewardAmount: logRecord.RewardAmount,
		CreatedAt:    logRecord.CreatedAt,
	}
}

func InviteRewardRecordFromRedeem(code *service.RedeemCode) *InviteRewardRecord {
	if code == nil {
		return nil
	}
	return &InviteRewardRecord{
		ID:        code.ID,
		Amount:    code.Value,
		UsedAt:    code.UsedAt,
		Notes:     code.Notes,
		CreatedAt: code.CreatedAt,
	}
}
