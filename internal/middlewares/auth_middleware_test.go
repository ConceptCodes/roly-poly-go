package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type mockUserRepo struct {
	findByApiKeyFunc func(ctx context.Context, apiKey string) (*models.UserModel, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.UserModel) error { return nil }
func (m *mockUserRepo) FindByApiKey(ctx context.Context, apiKey string) (*models.UserModel, error) {
	return m.findByApiKeyFunc(ctx, apiKey)
}

func TestAuthMiddlewareWhitelist(t *testing.T) {
	paths := []string{
		constants.HealthCheckEndpoint,
		constants.ReadinessEndpoint,
		constants.OnboardUserEndpoint,
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			var called bool
			middleware := NewAuthMiddleware(nil)
			middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})).ServeHTTP(rec, req)

			if !called {
				t.Errorf("whitelisted path %s was blocked", path)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d for whitelisted path %s", rec.Code, path)
			}
		})
	}
}

func TestAuthMiddlewareMissingKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware(nil)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ErrorCode != "RP-401" {
		t.Errorf("error_code = %s, want RP-401", resp.ErrorCode)
	}
}
