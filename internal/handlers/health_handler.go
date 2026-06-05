package handlers

import (
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
	"roly-poly/pkg/storage/postgres"
	redis2 "roly-poly/pkg/storage/redis"
)

type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

type Service struct {
	Name string
	Fn   func() bool
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redisClient: redisClient}
}

// ServiceAliveHandler godoc
// @Summary Check if service is alive
// @Description Check if service is alive
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/health/alive [get]
func (h *HealthHandler) ServiceAliveHandler(w http.ResponseWriter, r *http.Request) {
	helpers.SendSuccessResponse(w, "Service is alive", nil)
	return
}

// ServiceReadyHandler godoc
// @Summary Check if our services are ready
// @Description Check if service is ready
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/health/ready [get]
func (h *HealthHandler) ServiceReadyHandler(w http.ResponseWriter, r *http.Request) {

	services := []Service{
		{Name: "Postgres", Fn: func() bool { return postgres.HealthCheck(h.db) }},
	}

	if h.redisClient != nil {
		services = append(services, Service{
			Name: "Redis",
			Fn:   func() bool { return redis2.HealthCheck(r.Context(), h.redisClient) },
		})
	}

	var wg sync.WaitGroup
	responses := make([]models.HealthCheckResponseDto, len(services))

	wg.Add(len(services))

	for i, service := range services {
		go func(i int, service Service) {
			defer wg.Done()
			responses[i] = Runner(service.Name, service.Fn)
		}(i, service)
	}

	wg.Wait()

	helpers.SendSuccessResponse(w, "Service is ready", responses)
	return
}

func Runner(name string, fn func() bool) models.HealthCheckResponseDto {
	return models.HealthCheckResponseDto{
		Service: name,
		Status:  fn(),
	}
}
