package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

func TestServiceAliveHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health/alive", nil)
	rec := httptest.NewRecorder()

	handler := NewHealthHandler(nil)
	handler.ServiceAliveHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Message != "Service is alive" {
		t.Errorf("message = %q, want %q", resp.Message, "Service is alive")
	}
}

func TestRunner(t *testing.T) {
	tests := []struct {
		name   string
		status bool
	}{
		{"healthy", true},
		{"unhealthy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Runner(tt.name, func() bool { return tt.status })
			if result.Service != tt.name {
				t.Errorf("service name = %q, want %q", result.Service, tt.name)
			}
			if result.Status != tt.status {
				t.Errorf("status = %v, want %v", result.Status, tt.status)
			}
		})
	}
}

func TestServiceReadyHandler(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectPing()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health/status", nil)
	rec := httptest.NewRecorder()

	handler := NewHealthHandler(gormDB)
	handler.ServiceReadyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Message != "Service is ready" {
		t.Errorf("message = %q, want %q", resp.Message, "Service is ready")
	}

	services, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("response data is not an array")
	}
	if len(services) != 1 {
		t.Fatalf("len(services) = %d, want 1", len(services))
	}

	svc := services[0].(map[string]interface{})
	if svc["service"] != "Postgres" {
		t.Errorf("service name = %q, want %q", svc["service"], "Postgres")
	}
	if svc["status"] != true {
		t.Errorf("status = %v, want true", svc["status"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
