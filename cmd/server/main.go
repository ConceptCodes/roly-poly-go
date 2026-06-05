package server

// @title           Roly Poly API
// @version         1.0.0
// @description     REST API for creating polls, voting on options, and reporting results.
// @termsOfService  https://github.com/conceptcodes/roly-poly-go

// @contact.name   Roly Poly Support
// @contact.email  support@rolypoly.dev

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       x-api-key
// @description                API key obtained from the /api/onboard endpoint.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	swagger "github.com/swaggo/http-swagger/v2"

	_ "roly-poly/docs"

	"roly-poly/config"
	"roly-poly/internal/constants"
	"roly-poly/internal/handlers"
	"roly-poly/internal/middlewares"
	repository "roly-poly/internal/repositories"
	"roly-poly/pkg/logger"
	"roly-poly/pkg/ratelimit"
	"roly-poly/pkg/storage/postgres"
	redis2 "roly-poly/pkg/storage/redis"
	"roly-poly/pkg/tracer"
)

func Run() {
	log := logger.New()

	tp, err := tracer.InitTracerProvider(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize tracer provider, tracing disabled")
	}

	db, err := postgres.New()

	if err != nil {
		log.Fatal().Err(err).Msg("Error while connecting to database")
	}

	userRepo := repository.NewGormUserRepository(db)
	pollRepo := repository.NewGormPollRepository(db)
	optionRepo := repository.NewGormOptionRepository(db)

	redisClient, err := redis2.New(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available, rate limiting disabled")
	}

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	adminHandler := handlers.NewAdminHandler(userRepo)
	pollHandler := handlers.NewPollHandler(pollRepo, optionRepo)
	voteHandler := handlers.NewVoteHandler(pollRepo)

	router := mux.NewRouter()

	router.Use(middlewares.CORS(config.AppConfig.CorsAllowedOrigins))
	router.Use(middlewares.Recovery)
	router.Use(middlewares.BodyLimit(middlewares.MaxBodyBytes))
	router.Use(middlewares.TraceRequest)
	router.Use(middlewares.ContentTypeJSON)
	router.Use(middlewares.Tracing(config.AppConfig.OtelServiceName))
	router.Use(middlewares.Metrics)
	router.Use(middlewares.NewAuthMiddleware(userRepo))
	router.Use(middlewares.RequestLogger)
	router.NotFoundHandler = http.HandlerFunc(middlewares.NotFound)

	router.HandleFunc(constants.HealthCheckEndpoint, healthHandler.ServiceAliveHandler).Methods("GET")
	router.HandleFunc(constants.ReadinessEndpoint, healthHandler.ServiceReadyHandler).Methods("GET")
	router.Handle(constants.MetricsEndpoint, middlewares.MetricsHandler()).Methods("GET")
	router.PathPrefix(constants.DocsEndpoint).Handler(swagger.WrapHandler).Methods("GET")

	if redisClient != nil {
		onboardLimiter := middlewares.NewRateLimitMiddleware(
			ratelimit.New(redisClient, 10, 60*time.Second,
				ratelimit.WithKeyFunc(func(ip string) string { return fmt.Sprintf("ratelimit:onboard:%s", ip) }),
			),
		)
		router.Handle(constants.OnboardUserEndpoint, onboardLimiter(http.HandlerFunc(adminHandler.OnboardUser))).Methods("POST")

		pollLimiter := middlewares.NewRateLimitMiddleware(
			ratelimit.New(redisClient, 30, 60*time.Second,
				ratelimit.WithKeyFunc(func(ip string) string { return fmt.Sprintf("ratelimit:create_poll:%s", ip) }),
			),
		)
		router.Handle(constants.CreatePollEndpoint, pollLimiter(http.HandlerFunc(pollHandler.CreatePoll))).Methods("POST")

		voteLimiter := middlewares.NewRateLimitMiddleware(
			ratelimit.New(redisClient, 30, 60*time.Second,
				ratelimit.WithKeyFunc(func(ip string) string { return fmt.Sprintf("ratelimit:vote:%s", ip) }),
			),
		)
		router.Handle(constants.CastVoteEndpoint, voteLimiter(http.HandlerFunc(voteHandler.CastVote))).Methods("POST")
	} else {
		router.HandleFunc(constants.OnboardUserEndpoint, adminHandler.OnboardUser).Methods("POST")
		router.HandleFunc(constants.CreatePollEndpoint, pollHandler.CreatePoll).Methods("POST")
		router.HandleFunc(constants.CastVoteEndpoint, voteHandler.CastVote).Methods("POST")
	}

	router.HandleFunc(constants.ClosePollEndpoint, pollHandler.ClosePoll).Methods("PATCH")
	router.HandleFunc(constants.GetPollsEndpoint, pollHandler.GetPolls).Methods("GET")
	router.HandleFunc(constants.UpdatePollEndpoint, pollHandler.UpdatePoll).Methods("PATCH")
	router.HandleFunc(constants.DeletePollEndpoint, pollHandler.DeletePoll).Methods("DELETE")
	router.HandleFunc(constants.PollByIdEndpoint, pollHandler.GetPollById).Methods("GET")
	router.HandleFunc(constants.ReportEndpoint, pollHandler.GetPollReport).Methods("GET")

	port := fmt.Sprint(config.AppConfig.Port)
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: time.Duration(config.AppConfig.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.AppConfig.ReadTimeout) * time.Second,
	}

	go func() {
		log.Debug().Msgf(constants.StartMessage, port, config.AppConfig.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Error while starting server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	if redisClient != nil {
		redisClient.Close()
	}

	if tp != nil {
		_ = tp.Shutdown(context.Background())
	}

	postgres.Close(db)
	log.Info().Msg("Server exited")
}
