package service

import (
	"context"
	"encoding/json"
	"log"
)

type AdminActionLogRepository interface {
	Create(ctx context.Context, logRecord *AdminActionLog) error
}

type AdminActionLogService struct {
	repo AdminActionLogRepository
}

func NewAdminActionLogService(repo AdminActionLogRepository) *AdminActionLogService {
	return &AdminActionLogService{repo: repo}
}

func (s *AdminActionLogService) Log(ctx context.Context, input AdminActionLogInput) {
	if s == nil || s.repo == nil {
		return
	}
	payload := input.Payload
	if payload == "" {
		payload = "{}"
	}
	record := &AdminActionLog{
		AdminID:      input.AdminID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Payload:      payload,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		log.Printf("[AdminActionLog] failed to write log: %v", err)
	}
}

func MarshalAdminActionPayload(value any) string {
	if value == nil {
		return "{}"
	}
	b, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(b)
}
