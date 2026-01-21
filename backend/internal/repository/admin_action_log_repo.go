package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type adminActionLogRepository struct {
	client *ent.Client
}

func NewAdminActionLogRepository(client *ent.Client) service.AdminActionLogRepository {
	return &adminActionLogRepository{client: client}
}

func (r *adminActionLogRepository) Create(ctx context.Context, logRecord *service.AdminActionLog) error {
	if logRecord == nil {
		return nil
	}
	client := clientFromContext(ctx, r.client)
	created, err := client.AdminActionLog.Create().
		SetNillableAdminID(logRecord.AdminID).
		SetAction(logRecord.Action).
		SetResourceType(logRecord.ResourceType).
		SetNillableResourceID(logRecord.ResourceID).
		SetNillablePayload(nullableString(logRecord.Payload)).
		SetNillableIPAddress(nullableString(logRecord.IPAddress)).
		SetNillableUserAgent(nullableString(logRecord.UserAgent)).
		Save(ctx)
	if err != nil {
		return err
	}
	logRecord.ID = created.ID
	logRecord.CreatedAt = created.CreatedAt
	return nil
}
