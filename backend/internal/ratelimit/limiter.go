package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	redis *redis.Client
}

func New(redis *redis.Client) *Limiter {
	return &Limiter{redis: redis}
}

func (l *Limiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (allowed bool, remaining int, reset time.Time, err error) {

	pipe := l.redis.TxPipeline()

	count := pipe.Incr(ctx, key)
	ttl := pipe.TTL(ctx, key)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return true, 0, time.Time{}, err // 🔥 fail-open
	}

	// first hit → set expiry
	if ttl.Val() < 0 {
		_ = l.redis.Expire(ctx, key, window).Err()
		reset = time.Now().Add(window)
	} else {
		reset = time.Now().Add(ttl.Val())
	}

	remaining = limit - int(count.Val())
	allowed = count.Val() <= int64(limit)

	return
}
