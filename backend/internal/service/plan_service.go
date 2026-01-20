package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	Update(ctx context.Context, plan *Plan) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Plan, error)
	List(ctx context.Context, params pagination.PaginationParams, enabled *bool) ([]Plan, *pagination.PaginationResult, error)
}

type PlanService struct {
	planRepo PlanRepository
}

func NewPlanService(planRepo PlanRepository) *PlanService {
	return &PlanService{planRepo: planRepo}
}

func (s *PlanService) Create(ctx context.Context, input *CreatePlanInput) (*Plan, error) {
	if input == nil {
		return nil, nil
	}
	plan := &Plan{
		Title:         strings.TrimSpace(input.Title),
		Description:   strings.TrimSpace(input.Description),
		Price:         normalizePositiveFloat(input.Price),
		GroupName:     normalizeGroupName(input.GroupName),
		GroupSort:     input.GroupSort,
		DailyQuota:    normalizePositiveFloat(input.DailyQuota),
		TotalQuota:    normalizePositiveFloat(input.TotalQuota),
		PurchaseQRURL: strings.TrimSpace(input.PurchaseQRURL),
		Enabled:       input.Enabled,
		SortOrder:     input.SortOrder,
	}
	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Update(ctx context.Context, id int64, input *UpdatePlanInput) (*Plan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input == nil {
		return plan, nil
	}

	if input.Title != nil {
		plan.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		plan.Description = strings.TrimSpace(*input.Description)
	}
	if input.Price != nil {
		plan.Price = normalizePositiveFloat(*input.Price)
	}
	if input.GroupName != nil {
		plan.GroupName = normalizeGroupName(*input.GroupName)
	}
	if input.GroupSort != nil {
		plan.GroupSort = *input.GroupSort
	}
	if input.DailyQuota != nil {
		plan.DailyQuota = normalizePositiveFloat(*input.DailyQuota)
	}
	if input.TotalQuota != nil {
		plan.TotalQuota = normalizePositiveFloat(*input.TotalQuota)
	}
	if input.PurchaseQRURL != nil {
		plan.PurchaseQRURL = strings.TrimSpace(*input.PurchaseQRURL)
	}
	if input.Enabled != nil {
		plan.Enabled = *input.Enabled
	}
	if input.SortOrder != nil {
		plan.SortOrder = *input.SortOrder
	}

	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Delete(ctx context.Context, id int64) error {
	return s.planRepo.Delete(ctx, id)
}

func (s *PlanService) GetByID(ctx context.Context, id int64) (*Plan, error) {
	return s.planRepo.GetByID(ctx, id)
}

func (s *PlanService) List(ctx context.Context, params pagination.PaginationParams, enabled *bool) ([]Plan, *pagination.PaginationResult, error) {
	return s.planRepo.List(ctx, params, enabled)
}

func normalizeGroupName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	return value
}

func normalizePositiveFloat(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
