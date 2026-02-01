package repository

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/plan"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type planRepository struct {
	client *ent.Client
}

func NewPlanRepository(client *ent.Client) service.PlanRepository {
	return &planRepository{client: client}
}

func (r *planRepository) Create(ctx context.Context, planModel *service.Plan) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.Plan.Create().
		SetTitle(planModel.Title).
		SetNillableDescription(nullableString(planModel.Description)).
		SetPrice(planModel.Price).
		SetGroupName(planModel.GroupName).
		SetGroupSort(planModel.GroupSort).
		SetDailyQuota(planModel.DailyQuota).
		SetTotalQuota(planModel.TotalQuota).
		SetNillablePurchaseQrURL(nullableString(planModel.PurchaseQRURL)).
		SetEnabled(planModel.Enabled).
		SetSortOrder(planModel.SortOrder).
		Save(ctx)
	if err != nil {
		return err
	}
	applyPlanEntityToService(planModel, created)
	return nil
}

func (r *planRepository) Update(ctx context.Context, planModel *service.Plan) error {
	client := clientFromContext(ctx, r.client)
	updated, err := client.Plan.UpdateOneID(planModel.ID).
		SetTitle(planModel.Title).
		SetNillableDescription(nullableString(planModel.Description)).
		SetPrice(planModel.Price).
		SetGroupName(planModel.GroupName).
		SetGroupSort(planModel.GroupSort).
		SetDailyQuota(planModel.DailyQuota).
		SetTotalQuota(planModel.TotalQuota).
		SetNillablePurchaseQrURL(nullableString(planModel.PurchaseQRURL)).
		SetEnabled(planModel.Enabled).
		SetSortOrder(planModel.SortOrder).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return service.ErrPlanNotFound
		}
		return err
	}
	planModel.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *planRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.Plan.Delete().Where(plan.IDEQ(id)).Exec(ctx)
	if ent.IsNotFound(err) {
		return service.ErrPlanNotFound
	}
	return err
}

func (r *planRepository) GetByID(ctx context.Context, id int64) (*service.Plan, error) {
	model, err := r.client.Plan.Query().Where(plan.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, service.ErrPlanNotFound
		}
		return nil, err
	}
	return planEntityToService(model), nil
}

func (r *planRepository) List(ctx context.Context, params pagination.PaginationParams, enabled *bool) ([]service.Plan, *pagination.PaginationResult, error) {
	q := r.client.Plan.Query()
	if enabled != nil {
		q = q.Where(plan.EnabledEQ(*enabled))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	items, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(
			plan.ByGroupSort(sql.OrderAsc()),
			plan.ByGroupName(sql.OrderAsc()),
			plan.BySortOrder(sql.OrderAsc()),
			plan.ByID(sql.OrderAsc()),
		).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	plans := make([]service.Plan, 0, len(items))
	for i := range items {
		if v := planEntityToService(items[i]); v != nil {
			plans = append(plans, *v)
		}
	}

	return plans, paginationResultFromTotal(int64(total), params), nil
}

func (r *planRepository) UpdateGroupSorts(ctx context.Context, updates []service.PlanGroupSort) error {
	if len(updates) == 0 {
		return nil
	}

	client := clientFromContext(ctx, r.client)
	for _, update := range updates {
		if update.GroupName == "" {
			continue
		}
		_, err := client.Plan.Update().
			Where(plan.GroupNameEQ(update.GroupName)).
			SetGroupSort(update.GroupSort).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func planEntityToService(m *ent.Plan) *service.Plan {
	if m == nil {
		return nil
	}
	return &service.Plan{
		ID:            m.ID,
		Title:         m.Title,
		Description:   derefString(m.Description),
		Price:         m.Price,
		GroupName:     m.GroupName,
		GroupSort:     m.GroupSort,
		DailyQuota:    m.DailyQuota,
		TotalQuota:    m.TotalQuota,
		PurchaseQRURL: derefString(m.PurchaseQrURL),
		Enabled:       m.Enabled,
		SortOrder:     m.SortOrder,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func applyPlanEntityToService(target *service.Plan, m *ent.Plan) {
	if target == nil || m == nil {
		return
	}
	target.ID = m.ID
	target.CreatedAt = m.CreatedAt
	target.UpdatedAt = m.UpdatedAt
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
