package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitStore interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

type redisStore struct {
	client *redis.Client
}

func (s *redisStore) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

func (s *redisStore) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return s.client.Expire(ctx, key, expiration).Err()
}

type Limiter struct {
	client  RateLimitStore
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
		client:  &redisStore{client: client},
		limit:   limit,
		window:  window,
		keyFunc: func(ip string) string { return fmt.Sprintf("ratelimit:%s", ip) },
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func NewWithStore(store RateLimitStore, limit int, window time.Duration, opts ...Option) *Limiter {
	l := &Limiter{
		client:  store,
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

	count, err := l.client.Incr(ctx, key)
	if err != nil {
		return false, fmt.Errorf("rate limit check: %w", err)
	}

	if err := l.client.Expire(ctx, key, l.window); err != nil {
		return false, fmt.Errorf("rate limit set expiry: %w", err)
	}

	return count <= int64(l.limit), nil
}
