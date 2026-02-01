package service

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrPlanNotFound = infraerrors.NotFound("PLAN_NOT_FOUND", "plan not found")
)

type Plan struct {
	ID            int64
	Title         string
	Description   string
	Price         float64
	GroupName     string
	GroupSort     int
	DailyQuota    float64
	TotalQuota    float64
	PurchaseQRURL string
	Enabled       bool
	SortOrder     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreatePlanInput struct {
	Title         string
	Description   string
	Price         float64
	GroupName     string
	GroupSort     int
	DailyQuota    float64
	TotalQuota    float64
	PurchaseQRURL string
	Enabled       bool
	SortOrder     int
}

type UpdatePlanInput struct {
	Title         *string
	Description   *string
	Price         *float64
	GroupName     *string
	GroupSort     *int
	DailyQuota    *float64
	TotalQuota    *float64
	PurchaseQRURL *string
	Enabled       *bool
	SortOrder     *int
}

type PlanGroupSort struct {
	GroupName string
	GroupSort int
}
