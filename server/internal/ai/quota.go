package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultDailyLimit = 20
	quotaKeyPrefix    = "ai:quota"
)

type QuotaService struct {
	client    *redis.Client
	dailyLimit int
}

func NewQuotaService(redisAddr string) *QuotaService {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &QuotaService{client: client, dailyLimit: defaultDailyLimit}
}

func (q *QuotaService) key(userID string) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s:%s:%s", quotaKeyPrefix, userID, date)
}

// Check returns remaining quota and whether the user is allowed.
func (q *QuotaService) Check(userID string) (int, bool) {
	val, err := q.client.Get(context.Background(), q.key(userID)).Int()
	if err == redis.Nil {
		return q.dailyLimit, true
	}
	if err != nil {
		return 0, false
	}
	remaining := q.dailyLimit - val
	if remaining < 0 {
		remaining = 0
	}
	return remaining, remaining > 0
}

// Consume increments the usage counter. Returns error only on Redis failures.
func (q *QuotaService) Consume(userID string) error {
	key := q.key(userID)
	ctx := context.Background()
	n, err := q.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("quota incr: %w", err)
	}
	if n == 1 {
		// Set TTL to end of day
		now := time.Now()
		endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		q.client.ExpireAt(ctx, key, endOfDay)
	}
	return nil
}
