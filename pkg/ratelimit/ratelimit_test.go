package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockStore struct {
	mu       sync.Mutex
	counters map[string]int64
	expireCalls map[string]time.Duration
	incrErr  error
	expireErr error
}

func newMockStore() *mockStore {
	return &mockStore{
		counters:    make(map[string]int64),
		expireCalls: make(map[string]time.Duration),
	}
}

func (m *mockStore) Incr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.incrErr != nil {
		return 0, m.incrErr
	}
	m.counters[key]++
	return m.counters[key], nil
}

func (m *mockStore) Expire(ctx context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.expireErr != nil {
		return m.expireErr
	}
	m.expireCalls[key] = expiration
	return nil
}

func TestLimiterAllow_UnderLimit(t *testing.T) {
	store := newMockStore()
	limiter := NewWithStore(store, 10, time.Minute)

	for i := 1; i <= 10; i++ {
		allowed, err := limiter.Allow(context.Background(), "1.2.3.4")
		if err != nil {
			t.Errorf("unexpected error at call %d: %v", i, err)
		}
		if !allowed {
			t.Errorf("call %d should be allowed (under limit)", i)
		}
	}
}

func TestLimiterAllow_OverLimit(t *testing.T) {
	store := newMockStore()
	limiter := NewWithStore(store, 5, time.Minute)

	for i := 1; i <= 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "1.2.3.4")
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Errorf("call %d should be allowed (at limit)", i)
		}
	}

	// 6th call should be denied
	allowed, err := limiter.Allow(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Error("call 6 should be denied (over limit)")
	}
}

func TestLimiterAllow_DifferentIPs(t *testing.T) {
	store := newMockStore()
	limiter := NewWithStore(store, 2, time.Minute)

	allowed, _ := limiter.Allow(context.Background(), "1.1.1.1")
	if !allowed {
		t.Error("first call for 1.1.1.1 should be allowed")
	}

	allowed, _ = limiter.Allow(context.Background(), "2.2.2.2")
	if !allowed {
		t.Error("first call for 2.2.2.2 should be allowed")
	}

	allowed, _ = limiter.Allow(context.Background(), "1.1.1.1")
	if !allowed {
		t.Error("second call for 1.1.1.1 should be allowed")
	}

	allowed, _ = limiter.Allow(context.Background(), "1.1.1.1")
	if allowed {
		t.Error("third call for 1.1.1.1 should be denied")
	}

	allowed, _ = limiter.Allow(context.Background(), "2.2.2.2")
	if !allowed {
		t.Error("second call for 2.2.2.2 should be allowed (at limit)")
	}

	allowed, _ = limiter.Allow(context.Background(), "2.2.2.2")
	if allowed {
		t.Error("third call for 2.2.2.2 should be denied (over limit)")
	}
}

func TestLimiterAllow_IncrError(t *testing.T) {
	store := newMockStore()
	store.incrErr = fmt.Errorf("redis error")
	limiter := NewWithStore(store, 10, time.Minute)

	_, err := limiter.Allow(context.Background(), "1.2.3.4")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestLimiterAllow_ExpireError(t *testing.T) {
	store := newMockStore()
	store.expireErr = fmt.Errorf("expire error")
	limiter := NewWithStore(store, 10, time.Minute)

	_, err := limiter.Allow(context.Background(), "1.2.3.4")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestWithKeyFunc(t *testing.T) {
	store := newMockStore()
	limiter := NewWithStore(store, 10, time.Minute,
		WithKeyFunc(func(ip string) string {
			return "custom:prefix:" + ip
		}),
	)

	limiter.Allow(context.Background(), "1.2.3.4")

	store.mu.Lock()
	key := "custom:prefix:1.2.3.4"
	count, ok := store.counters[key]
	store.mu.Unlock()

	if !ok {
		t.Errorf("expected counter for key %q to exist", key)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestDefaultKeyFunc(t *testing.T) {
	store := newMockStore()
	limiter := NewWithStore(store, 10, time.Minute)

	limiter.Allow(context.Background(), "1.2.3.4")

	store.mu.Lock()
	key := "ratelimit:1.2.3.4"
	count, ok := store.counters[key]
	store.mu.Unlock()

	if !ok {
		t.Errorf("expected counter for key %q to exist", key)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
