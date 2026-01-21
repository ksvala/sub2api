package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type Plan struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	Price         float64   `json:"price"`
	GroupName     string    `json:"group_name"`
	GroupSort     int       `json:"group_sort"`
	DailyQuota    float64   `json:"daily_quota"`
	TotalQuota    float64   `json:"total_quota"`
	PurchaseQRURL string    `json:"purchase_qr_url,omitempty"`
	Enabled       bool      `json:"enabled"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PublicPlan struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Price         float64 `json:"price"`
	GroupName     string  `json:"group_name"`
	GroupSort     int     `json:"group_sort"`
	DailyQuota    float64 `json:"daily_quota"`
	TotalQuota    float64 `json:"total_quota"`
	PurchaseQRURL string  `json:"purchase_qr_url,omitempty"`
}

func PlanFromService(plan *service.Plan) *Plan {
	if plan == nil {
		return nil
	}
	return &Plan{
		ID:            plan.ID,
		Title:         plan.Title,
		Description:   plan.Description,
		Price:         plan.Price,
		GroupName:     plan.GroupName,
		GroupSort:     plan.GroupSort,
		DailyQuota:    plan.DailyQuota,
		TotalQuota:    plan.TotalQuota,
		PurchaseQRURL: plan.PurchaseQRURL,
		Enabled:       plan.Enabled,
		SortOrder:     plan.SortOrder,
		CreatedAt:     plan.CreatedAt,
		UpdatedAt:     plan.UpdatedAt,
	}
}

func PublicPlanFromService(plan *service.Plan) *PublicPlan {
	if plan == nil {
		return nil
	}
	return &PublicPlan{
		ID:            plan.ID,
		Title:         plan.Title,
		Description:   plan.Description,
		Price:         plan.Price,
		GroupName:     plan.GroupName,
		GroupSort:     plan.GroupSort,
		DailyQuota:    plan.DailyQuota,
		TotalQuota:    plan.TotalQuota,
		PurchaseQRURL: plan.PurchaseQRURL,
	}
}
