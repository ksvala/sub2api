package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// SubscriptionReminderService sends usage/expiry reminders to users.
//
// Current strategy:
// - Triggered on user traffic (e.g. /subscriptions/active, /subscriptions/summary)
// - Email reminders are deduped via Redis SETNX keys
// - In-app reminders are handled on the frontend
type SubscriptionReminderService struct {
	userRepo          UserRepository
	userSubRepo       UserSubscriptionRepository
	redis             *redis.Client
	emailQueueService *EmailQueueService
}

func NewSubscriptionReminderService(
	userRepo UserRepository,
	userSubRepo UserSubscriptionRepository,
	redisClient *redis.Client,
	emailQueueService *EmailQueueService,
) *SubscriptionReminderService {
	return &SubscriptionReminderService{
		userRepo:          userRepo,
		userSubRepo:       userSubRepo,
		redis:             redisClient,
		emailQueueService: emailQueueService,
	}
}

func (s *SubscriptionReminderService) CheckAndNotify(ctx context.Context, userID int64) {
	if s == nil || s.userRepo == nil || s.userSubRepo == nil || s.redis == nil || s.emailQueueService == nil {
		return
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return
	}
	if user == nil || user.Email == "" {
		return
	}

	subs, err := s.userSubRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return
	}

	for i := range subs {
		s.checkAndNotifyOne(ctx, user.Email, &subs[i])
	}
}

func (s *SubscriptionReminderService) checkAndNotifyOne(ctx context.Context, email string, sub *UserSubscription) {
	if sub == nil || sub.Group == nil {
		return
	}

	// 1) Daily usage thresholds
	if sub.Group.HasDailyLimit() && sub.DailyWindowStart != nil {
		limit := *sub.Group.DailyLimitUSD
		used := sub.DailyUsageUSD
		if limit > 0 {
			pct := (used / limit) * 100
			dateKey := sub.DailyWindowStart.Format("2006-01-02")

			sent95 := s.tryMarkOnce(ctx, fmt.Sprintf("reminder:sub:%d:daily95:%s", sub.ID, dateKey), 36*time.Hour)
			if pct >= 95 && sent95 {
				s.sendEmailSafe(ctx, email, "订阅用量提醒：今日用量已达 95%", buildUsageEmailBody(sub, used, limit, pct))
				return
			}

			sent80 := s.tryMarkOnce(ctx, fmt.Sprintf("reminder:sub:%d:daily80:%s", sub.ID, dateKey), 36*time.Hour)
			if pct >= 80 && sent80 {
				s.sendEmailSafe(ctx, email, "订阅用量提醒：今日用量已达 80%", buildUsageEmailBody(sub, used, limit, pct))
			}
		}
	}

	// 2) Expiry reminders (3/1/0 days)
	days := sub.DaysRemaining()
	if days == 3 || days == 1 || days == 0 {
		key := fmt.Sprintf("reminder:sub:%d:expiry:%d", sub.ID, days)
		if s.tryMarkOnce(ctx, key, 36*time.Hour) {
			subject := fmt.Sprintf("订阅到期提醒：剩余 %d 天", days)
			if days == 0 {
				subject = "订阅到期提醒：今天到期"
			}
			s.sendEmailSafe(ctx, email, subject, buildExpiryEmailBody(sub, days))
		}
	}
}

func (s *SubscriptionReminderService) tryMarkOnce(ctx context.Context, key string, ttl time.Duration) bool {
	if key == "" {
		return false
	}
	ok, err := s.redis.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		log.Printf("[Reminder] setnx failed: key=%s err=%v", key, err)
		return false
	}
	return ok
}

func (s *SubscriptionReminderService) sendEmailSafe(ctx context.Context, email, subject, body string) {
	if email == "" || subject == "" || body == "" {
		return
	}
	if err := s.emailQueueService.EnqueueEmail(email, subject, body); err != nil {
		log.Printf("[Reminder] enqueue email failed: email=%s err=%v", email, err)
	}
}

func buildUsageEmailBody(sub *UserSubscription, used, limit, pct float64) string {
	groupName := ""
	if sub.Group != nil {
		groupName = sub.Group.Name
	}
	expiresAt := sub.ExpiresAt.Format(time.RFC3339)
	return fmt.Sprintf(
		"<p>你的订阅（%s）今日用量已达 <b>%.0f%%</b>。</p><p>已用：$%.2f / 限额：$%.2f</p><p>到期时间：%s</p>",
		groupName,
		pct,
		used,
		limit,
		expiresAt,
	)
}

func buildExpiryEmailBody(sub *UserSubscription, days int) string {
	groupName := ""
	if sub.Group != nil {
		groupName = sub.Group.Name
	}
	expiresAt := sub.ExpiresAt.Format(time.RFC3339)
	msg := fmt.Sprintf("<p>你的订阅（%s）距离到期还有 <b>%d</b> 天。</p>", groupName, days)
	if days == 0 {
		msg = fmt.Sprintf("<p>你的订阅（%s）<b>今天到期</b>。</p>", groupName)
	}
	return msg + fmt.Sprintf("<p>到期时间：%s</p>", expiresAt)
}
