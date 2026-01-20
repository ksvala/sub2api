package service

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrInviteNotFound        = infraerrors.NotFound("INVITE_NOT_FOUND", "invite not found")
	ErrInviteAlreadyBound    = infraerrors.Conflict("INVITE_ALREADY_BOUND", "invite already bound")
	ErrInviteAlreadyConfirmed = infraerrors.Conflict("INVITE_ALREADY_CONFIRMED", "invite already confirmed")
	ErrInviteCodeInvalid     = infraerrors.BadRequest("INVITE_CODE_INVALID", "invalid invite code")
)

type Invite struct {
	ID           int64
	InviterID    int64
	InviteeID    int64
	InviteCode   string
	RewardAmount float64
	Status       string
	ConfirmedBy  *int64
	ConfirmedAt  *time.Time
	CreatedAt    time.Time

	Inviter        *User
	Invitee        *User
	ConfirmedByUser *User
}

type InviteLog struct {
	ID           int64
	InviteID     int64
	Action       string
	InviterID    int64
	InviteeID    int64
	AdminID      *int64
	RewardAmount float64
	CreatedAt    time.Time

	Inviter *User
	Invitee *User
	Admin   *User
}

type InviteSummary struct {
	InviteCode        string
	TotalInvites      int
	PendingInvites    int
	ConfirmedInvites  int
	TotalRewardAmount float64
}

type InviteSettings struct {
	RewardAmount float64
}

type InviteLogFilters struct {
	Action       string
	InviterID    *int64
	InviteeID    *int64
	InviterEmail string
	InviteeEmail string
	StartTime    *time.Time
	EndTime      *time.Time
}
