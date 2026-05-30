package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	client  *redis.Client
	limit   int
	window  time.Duration
	keyFunc func(ip string) string
}

type Option func(*Limiter)

func WithKeyFunc(fn func(ip string) string) Option {
	return func(l *Limiter) {
		l.keyFunc = fn
	}
}

func New(client *redis.Client, limit int, window time.Duration, opts ...Option) *Limiter {
	l := &Limiter{
		client:  client,
		limit:   limit,
		window:  window,
		keyFunc: func(ip string) string { return fmt.Sprintf("ratelimit:%s", ip) },
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Limiter) Allow(ctx context.Context, ip string) (bool, error) {
	key := l.keyFunc(ip)
	pipe := l.client.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, l.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit check: %w", err)
	}

	return incr.Val() <= int64(l.limit), nil
}
