package handlers

import (
	"net/http"
	"sync"

	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
	"roly-poly/pkg/storage/postgres"
)

type HealthHandler struct{}

type Service struct {
	Name string
	Fn   func() bool
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
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
		{Name: "Postgres", Fn: postgres.HealthCheck},
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
		Up:      fn(),
	}
}
