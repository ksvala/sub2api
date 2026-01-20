package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	inviteCodeLength   = 6
	inviteCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *Invite) error
	Update(ctx context.Context, invite *Invite) error
	GetByInviteeID(ctx context.Context, inviteeID int64) (*Invite, error)
	GetByInviteeIDForUpdate(ctx context.Context, inviteeID int64) (*Invite, error)
	ListByInviter(ctx context.Context, inviterID int64, params pagination.PaginationParams, status string) ([]Invite, *pagination.PaginationResult, error)
	GetSummaryByInviter(ctx context.Context, inviterID int64) (int, int, int, float64, error)
}

type InviteLogRepository interface {
	Create(ctx context.Context, logRecord *InviteLog) error
	List(ctx context.Context, params pagination.PaginationParams, filters InviteLogFilters) ([]InviteLog, *pagination.PaginationResult, error)
}

type InviteService struct {
	entClient            *ent.Client
	inviteRepo           InviteRepository
	inviteLogRepo        InviteLogRepository
	userRepo             UserRepository
	redeemRepo           RedeemCodeRepository
	settingService       *SettingService
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

func NewInviteService(
	entClient *ent.Client,
	inviteRepo InviteRepository,
	inviteLogRepo InviteLogRepository,
	userRepo UserRepository,
	redeemRepo RedeemCodeRepository,
	settingService *SettingService,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *InviteService {
	return &InviteService{
		entClient:            entClient,
		inviteRepo:           inviteRepo,
		inviteLogRepo:        inviteLogRepo,
		userRepo:             userRepo,
		redeemRepo:           redeemRepo,
		settingService:       settingService,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
	}
}

func (s *InviteService) ResolveInviter(ctx context.Context, code string) (*User, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrInviteCodeInvalid
	}
	inviter, err := s.userRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, ErrInviteCodeInvalid
	}
	if inviter == nil {
		return nil, ErrInviteCodeInvalid
	}
	return inviter, nil
}

func (s *InviteService) GetOrCreateInviteCode(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(user.InviteCode) != "" {
		return user.InviteCode, nil
	}
	code, err := s.generateUniqueInviteCode(ctx)
	if err != nil {
		return "", err
	}
	if err := s.userRepo.SetInviteCode(ctx, userID, code); err != nil {
		return "", err
	}
	return code, nil
}

func (s *InviteService) BindInvite(ctx context.Context, inviterID, inviteeID int64, inviteCode string) (*Invite, error) {
	if inviterID == inviteeID {
		return nil, ErrInviteCodeInvalid
	}
	rewardAmount := s.getInviteRewardAmount(ctx)

	invite := &Invite{
		InviterID:    inviterID,
		InviteeID:    inviteeID,
		InviteCode:   inviteCode,
		RewardAmount: rewardAmount,
		Status:       InviteStatusPending,
	}

	if s.entClient == nil {
		if err := s.inviteRepo.Create(ctx, invite); err != nil {
			return nil, err
		}
		if err := s.inviteLogRepo.Create(ctx, &InviteLog{
			InviteID:     invite.ID,
			Action:       InviteLogActionBind,
			InviterID:    inviterID,
			InviteeID:    inviteeID,
			RewardAmount: rewardAmount,
			CreatedAt:    time.Now(),
		}); err != nil {
			return nil, err
		}
		return invite, nil
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := ent.NewTxContext(ctx, tx)
	if err := s.inviteRepo.Create(txCtx, invite); err != nil {
		return nil, err
	}
	if err := s.inviteLogRepo.Create(txCtx, &InviteLog{
		InviteID:     invite.ID,
		Action:       InviteLogActionBind,
		InviterID:    inviterID,
		InviteeID:    inviteeID,
		RewardAmount: rewardAmount,
		CreatedAt:    time.Now(),
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return invite, nil
}

func (s *InviteService) ConfirmInvite(ctx context.Context, inviteeID, adminID int64) (*Invite, error) {
	if s.entClient == nil {
		return nil, fmt.Errorf("ent client not configured")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := ent.NewTxContext(ctx, tx)
	invite, err := s.inviteRepo.GetByInviteeIDForUpdate(txCtx, inviteeID)
	if err != nil {
		return nil, err
	}
	if invite.Status == InviteStatusConfirmed {
		return nil, ErrInviteAlreadyConfirmed
	}

	now := time.Now()
	invite.Status = InviteStatusConfirmed
	invite.ConfirmedBy = &adminID
	invite.ConfirmedAt = &now
	if err := s.inviteRepo.Update(txCtx, invite); err != nil {
		return nil, err
	}

	if invite.RewardAmount != 0 {
		if err := s.userRepo.UpdateBalance(txCtx, invite.InviterID, invite.RewardAmount); err != nil {
			return nil, err
		}
	}

	code, err := GenerateRedeemCode()
	if err != nil {
		return nil, fmt.Errorf("generate invite reward code: %w", err)
	}
	adjustmentRecord := &RedeemCode{
		Code:   code,
		Type:   AdjustmentTypeInviteReward,
		Value:  invite.RewardAmount,
		Status: StatusUsed,
		UsedBy: &invite.InviterID,
		Notes:  fmt.Sprintf("Invite reward for user %d", invite.InviteeID),
	}
	adjustmentRecord.UsedAt = &now
	if err := s.redeemRepo.Create(txCtx, adjustmentRecord); err != nil {
		return nil, err
	}

	if err := s.inviteLogRepo.Create(txCtx, &InviteLog{
		InviteID:     invite.ID,
		Action:       InviteLogActionConfirm,
		InviterID:    invite.InviterID,
		InviteeID:    invite.InviteeID,
		AdminID:      &adminID,
		RewardAmount: invite.RewardAmount,
		CreatedAt:    now,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.invalidateInviteCaches(invite.InviterID, invite.RewardAmount)
	return invite, nil
}

func (s *InviteService) GetInviteSummary(ctx context.Context, userID int64) (*InviteSummary, error) {
	inviteCode, err := s.GetOrCreateInviteCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.inviteRepo == nil {
		return &InviteSummary{InviteCode: inviteCode}, nil
	}
	total, pending, confirmed, rewardSum, err := s.inviteRepo.GetSummaryByInviter(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &InviteSummary{
		InviteCode:        inviteCode,
		TotalInvites:      total,
		PendingInvites:    pending,
		ConfirmedInvites:  confirmed,
		TotalRewardAmount: rewardSum,
	}, nil
}

func (s *InviteService) ListInvitesByInviter(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]Invite, *pagination.PaginationResult, error) {
	if s.inviteRepo == nil {
		return []Invite{}, emptyInvitePagination(params), nil
	}
	return s.inviteRepo.ListByInviter(ctx, userID, params, status)
}

func (s *InviteService) ListInviteLogs(ctx context.Context, params pagination.PaginationParams, filters InviteLogFilters) ([]InviteLog, *pagination.PaginationResult, error) {
	if s.inviteLogRepo == nil {
		return []InviteLog{}, emptyInvitePagination(params), nil
	}
	return s.inviteLogRepo.List(ctx, params, filters)
}

func (s *InviteService) ListInviteRewardRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	if s.redeemRepo == nil {
		return []RedeemCode{}, emptyInvitePagination(params), nil
	}
	return s.redeemRepo.ListByUserWithFilters(ctx, userID, params, AdjustmentTypeInviteReward)
}

func (s *InviteService) GetInviteSettings(ctx context.Context) (*InviteSettings, error) {
	return &InviteSettings{RewardAmount: s.getInviteRewardAmount(ctx)}, nil
}

func (s *InviteService) UpdateInviteSettings(ctx context.Context, settings InviteSettings) (*InviteSettings, error) {
	if s.settingService == nil {
		return &InviteSettings{RewardAmount: settings.RewardAmount}, nil
	}
	if err := s.settingService.SetInviteRewardAmount(ctx, settings.RewardAmount); err != nil {
		return nil, err
	}
	return &InviteSettings{RewardAmount: settings.RewardAmount}, nil
}

func (s *InviteService) generateUniqueInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		code, err := randomInviteCode()
		if err != nil {
			return "", err
		}
		exists, err := s.userRepo.ExistsByInviteCode(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique invite code")
}

func randomInviteCode() (string, error) {
	b := make([]byte, inviteCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := 0; i < inviteCodeLength; i++ {
		b[i] = inviteCodeAlphabet[int(b[i])%len(inviteCodeAlphabet)]
	}
	return string(b), nil
}

func (s *InviteService) getInviteRewardAmount(ctx context.Context) float64 {
	if s.settingService == nil {
		return 0
	}
	return s.settingService.GetInviteRewardAmount(ctx)
}

func (s *InviteService) invalidateInviteCaches(userID int64, rewardAmount float64) {
	if s.authCacheInvalidator != nil && rewardAmount != 0 {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(context.Background(), userID)
	}
	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.billingCacheService.InvalidateUserBalance(cacheCtx, userID); err != nil {
				log.Printf("invalidate invite reward balance cache failed: user_id=%d err=%v", userID, err)
			}
		}()
	}
}

func emptyInvitePagination(params pagination.PaginationParams) *pagination.PaginationResult {
	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return &pagination.PaginationResult{
		Total:    0,
		Page:     page,
		PageSize: pageSize,
		Pages:    0,
	}
}
