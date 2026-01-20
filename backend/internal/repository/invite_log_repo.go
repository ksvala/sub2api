package repository

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invitelog"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteLogRepository struct {
	client *ent.Client
}

func NewInviteLogRepository(client *ent.Client) service.InviteLogRepository {
	return &inviteLogRepository{client: client}
}

func (r *inviteLogRepository) Create(ctx context.Context, logRecord *service.InviteLog) error {
	if logRecord == nil {
		return nil
	}
	client := clientFromContext(ctx, r.client)
	created, err := client.InviteLog.Create().
		SetInviteID(logRecord.InviteID).
		SetAction(logRecord.Action).
		SetInviterID(logRecord.InviterID).
		SetInviteeID(logRecord.InviteeID).
		SetNillableAdminID(logRecord.AdminID).
		SetRewardAmount(logRecord.RewardAmount).
		SetCreatedAt(logRecord.CreatedAt).
		Save(ctx)
	if err != nil {
		return err
	}
	logRecord.ID = created.ID
	logRecord.CreatedAt = created.CreatedAt
	return nil
}

func (r *inviteLogRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.InviteLogFilters) ([]service.InviteLog, *pagination.PaginationResult, error) {
	q := r.client.InviteLog.Query()
	if filters.Action != "" {
		q = q.Where(invitelog.ActionEQ(filters.Action))
	}
	if filters.InviterID != nil {
		q = q.Where(invitelog.InviterIDEQ(*filters.InviterID))
	}
	if filters.InviteeID != nil {
		q = q.Where(invitelog.InviteeIDEQ(*filters.InviteeID))
	}
	if filters.StartTime != nil {
		q = q.Where(invitelog.CreatedAtGTE(*filters.StartTime))
	}
	if filters.EndTime != nil {
		q = q.Where(invitelog.CreatedAtLTE(*filters.EndTime))
	}
	if filters.InviterEmail != "" {
		q = q.Where(invitelog.HasInviterWith(user.EmailContainsFold(filters.InviterEmail)))
	}
	if filters.InviteeEmail != "" {
		q = q.Where(invitelog.HasInviteeWith(user.EmailContainsFold(filters.InviteeEmail)))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	logs, err := q.
		WithInviter().
		WithInvitee().
		WithAdmin().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(invitelog.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := make([]service.InviteLog, 0, len(logs))
	for i := range logs {
		if v := inviteLogEntityToService(logs[i]); v != nil {
			out = append(out, *v)
		}
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func inviteLogEntityToService(m *ent.InviteLog) *service.InviteLog {
	if m == nil {
		return nil
	}
	out := &service.InviteLog{
		ID:           m.ID,
		InviteID:     m.InviteID,
		Action:       m.Action,
		InviterID:    m.InviterID,
		InviteeID:    m.InviteeID,
		AdminID:      m.AdminID,
		RewardAmount: m.RewardAmount,
		CreatedAt:    m.CreatedAt,
	}
	if m.Edges.Inviter != nil {
		out.Inviter = userEntityToService(m.Edges.Inviter)
	}
	if m.Edges.Invitee != nil {
		out.Invitee = userEntityToService(m.Edges.Invitee)
	}
	if m.Edges.Admin != nil {
		out.Admin = userEntityToService(m.Edges.Admin)
	}
	return out
}
