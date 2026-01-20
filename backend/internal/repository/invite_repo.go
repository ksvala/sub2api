package repository

import (
	"context"
	stdsql "database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invitation"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteRepository struct {
	client *ent.Client
}

func NewInviteRepository(client *ent.Client) service.InviteRepository {
	return &inviteRepository{client: client}
}

func (r *inviteRepository) Create(ctx context.Context, invite *service.Invite) error {
	if invite == nil {
		return nil
	}
	client := clientFromContext(ctx, r.client)
	created, err := client.Invitation.Create().
		SetInviterID(invite.InviterID).
		SetInviteeID(invite.InviteeID).
		SetInviteCode(invite.InviteCode).
		SetRewardAmount(invite.RewardAmount).
		SetStatus(invite.Status).
		SetNillableConfirmedBy(invite.ConfirmedBy).
		SetNillableConfirmedAt(invite.ConfirmedAt).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrInviteAlreadyBound)
	}
	invite.ID = created.ID
	invite.CreatedAt = created.CreatedAt
	return nil
}

func (r *inviteRepository) Update(ctx context.Context, invite *service.Invite) error {
	if invite == nil {
		return nil
	}
	client := clientFromContext(ctx, r.client)
	_, err := client.Invitation.UpdateOneID(invite.ID).
		SetStatus(invite.Status).
		SetRewardAmount(invite.RewardAmount).
		SetInviteCode(invite.InviteCode).
		SetNillableConfirmedBy(invite.ConfirmedBy).
		SetNillableConfirmedAt(invite.ConfirmedAt).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrInviteNotFound, nil)
	}
	return nil
}

func (r *inviteRepository) GetByInviteeID(ctx context.Context, inviteeID int64) (*service.Invite, error) {
	entity, err := r.client.Invitation.Query().
		Where(invitation.InviteeIDEQ(inviteeID)).
		WithInviter().
		WithInvitee().
		WithConfirmedByUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrInviteNotFound, nil)
	}
	return inviteEntityToService(entity), nil
}

func (r *inviteRepository) GetByInviteeIDForUpdate(ctx context.Context, inviteeID int64) (*service.Invite, error) {
	client := clientFromContext(ctx, r.client)
	entity, err := client.Invitation.Query().
		Where(invitation.InviteeIDEQ(inviteeID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrInviteNotFound, nil)
	}
	return inviteEntityToService(entity), nil
}

func (r *inviteRepository) ListByInviter(ctx context.Context, inviterID int64, params pagination.PaginationParams, status string) ([]service.Invite, *pagination.PaginationResult, error) {
	q := r.client.Invitation.Query().Where(invitation.InviterIDEQ(inviterID))
	if status != "" {
		q = q.Where(invitation.StatusEQ(status))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	invites, err := q.
		WithInvitee().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(invitation.ByCreatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	outInvites := make([]service.Invite, 0, len(invites))
	for i := range invites {
		if v := inviteEntityToService(invites[i]); v != nil {
			outInvites = append(outInvites, *v)
		}
	}
	return outInvites, paginationResultFromTotal(int64(total), params), nil
}

func (r *inviteRepository) GetSummaryByInviter(ctx context.Context, inviterID int64) (int, int, int, float64, error) {
	base := r.client.Invitation.Query().Where(invitation.InviterIDEQ(inviterID))
	total, err := base.Clone().Count(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	pending, err := base.Clone().Where(invitation.StatusEQ(service.InviteStatusPending)).Count(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	confirmed, err := base.Clone().Where(invitation.StatusEQ(service.InviteStatusConfirmed)).Count(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if confirmed == 0 {
		return total, pending, confirmed, 0, nil
	}
	var rewardSum stdsql.NullFloat64
	if err := base.Clone().
		Where(invitation.StatusEQ(service.InviteStatusConfirmed)).
		Aggregate(ent.Sum(invitation.FieldRewardAmount)).
		Scan(ctx, &rewardSum); err != nil {
		return 0, 0, 0, 0, err
	}

	if rewardSum.Valid {
		return total, pending, confirmed, rewardSum.Float64, nil
	}
	return total, pending, confirmed, 0, nil
}

func inviteEntityToService(m *ent.Invitation) *service.Invite {
	if m == nil {
		return nil
	}
	out := &service.Invite{
		ID:           m.ID,
		InviterID:    m.InviterID,
		InviteeID:    m.InviteeID,
		InviteCode:   m.InviteCode,
		RewardAmount: m.RewardAmount,
		Status:       m.Status,
		ConfirmedBy:  m.ConfirmedBy,
		ConfirmedAt:  m.ConfirmedAt,
		CreatedAt:    m.CreatedAt,
	}
	if m.Edges.Inviter != nil {
		out.Inviter = userEntityToService(m.Edges.Inviter)
	}
	if m.Edges.Invitee != nil {
		out.Invitee = userEntityToService(m.Edges.Invitee)
	}
	if m.Edges.ConfirmedByUser != nil {
		out.ConfirmedByUser = userEntityToService(m.Edges.ConfirmedByUser)
	}
	return out
}
