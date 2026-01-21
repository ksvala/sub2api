package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// BanLimiter tracks repeated failures and temporarily bans an IP.
type BanLimiter struct {
	redis  *redis.Client
	prefix string
}

// NewBanLimiter creates a BanLimiter instance.
func NewBanLimiter(redisClient *redis.Client) *BanLimiter {
	return &BanLimiter{redis: redisClient, prefix: "ban:"}
}

// BanOnFailure blocks banned IPs and records failures for the given scope.
func (b *BanLimiter) BanOnFailure(scope string, threshold int, window time.Duration, banDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if b == nil || b.redis == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		banned, ttl, err := b.isBanned(c.Request.Context(), scope, ip)
		if err != nil {
			log.Printf("[BanLimiter] check failed: scope=%s ip=%s err=%v", scope, ip, err)
			c.Next()
			return
		}
		if banned {
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "banned",
				"message": "Too many failed attempts, please try again later",
			})
			return
		}

		c.Next()

		status := c.Writer.Status()
		if status >= 400 && status < 500 && status != http.StatusTooManyRequests {
			if _, err := b.recordFailure(c.Request.Context(), scope, ip, threshold, window, banDuration); err != nil {
				log.Printf("[BanLimiter] record failure: scope=%s ip=%s err=%v", scope, ip, err)
			}
		}
	}
}

func (b *BanLimiter) isBanned(ctx context.Context, scope string, ip string) (bool, time.Duration, error) {
	banKey := b.banKey(scope, ip)
	pttl, err := b.redis.PTTL(ctx, banKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, 0, nil
		}
		return false, 0, err
	}
	if pttl > 0 {
		return true, pttl, nil
	}
	return false, 0, nil
}

var banScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
local ttl = redis.call('PTTL', KEYS[1])
if count == 1 or ttl == -1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
if count >= tonumber(ARGV[2]) then
  redis.call('SET', KEYS[2], '1', 'PX', ARGV[3])
  redis.call('DEL', KEYS[1])
  return {count, 1}
end
return {count, 0}
`)

func (b *BanLimiter) recordFailure(ctx context.Context, scope string, ip string, threshold int, window time.Duration, banDuration time.Duration) (bool, error) {
	key := b.counterKey(scope, ip)
	banKey := b.banKey(scope, ip)

	values, err := banScript.Run(ctx, b.redis, []string{key, banKey}, window.Milliseconds(), threshold, banDuration.Milliseconds()).Slice()
	if err != nil {
		return false, err
	}
	if len(values) < 2 {
		return false, fmt.Errorf("ban limiter script returned %d values", len(values))
	}
	flag, ok := values[1].(int64)
	if ok && flag == 1 {
		return true, nil
	}
	if s, ok := values[1].(string); ok && s == "1" {
		return true, nil
	}
	return false, nil
}

func (b *BanLimiter) counterKey(scope, ip string) string {
	return b.prefix + "counter:" + sanitizeScope(scope) + ":" + ip
}

func (b *BanLimiter) banKey(scope, ip string) string {
	return b.prefix + "active:" + sanitizeScope(scope) + ":" + ip
}

func sanitizeScope(scope string) string {
	return strings.ReplaceAll(scope, " ", "-")
}
