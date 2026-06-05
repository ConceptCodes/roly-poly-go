package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"roly-poly/config"
	"roly-poly/internal/constants"
	"roly-poly/internal/handlers"
	"roly-poly/internal/middlewares"
	repository "roly-poly/internal/repositories"
	"roly-poly/pkg/logger"
	"roly-poly/pkg/ratelimit"
	"roly-poly/pkg/storage/postgres"
	redis2 "roly-poly/pkg/storage/redis"
)

func Run() {
	log := logger.New()

	db, err := postgres.New()

	if err != nil {
		log.Fatal().Err(err).Msg("Error while connecting to database")
	}

	userRepo := repository.NewGormUserRepository(db)
	pollRepo := repository.NewGormPollRepository(db)
	optionRepo := repository.NewGormOptionRepository(db)

	healthHandler := handlers.NewHealthHandler(db)
	adminHandler := handlers.NewAdminHandler(userRepo)
	pollHandler := handlers.NewPollHandler(pollRepo, optionRepo)
	voteHandler := handlers.NewVoteHandler(pollRepo)

	redisClient, err := redis2.New(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available, rate limiting disabled")
	}

	router := mux.NewRouter()

	router.Use(middlewares.Recovery)
	router.Use(middlewares.BodyLimit(middlewares.MaxBodyBytes))
	router.Use(middlewares.TraceRequest)
	router.Use(middlewares.ContentTypeJSON)
	router.Use(middlewares.NewAuthMiddleware(userRepo))
	router.Use(middlewares.RequestLogger)
	router.NotFoundHandler = http.HandlerFunc(middlewares.NotFound)

	router.HandleFunc(constants.HealthCheckEndpoint, healthHandler.ServiceAliveHandler).Methods("GET")
	router.HandleFunc(constants.ReadinessEndpoint, healthHandler.ServiceReadyHandler).Methods("GET")

	if redisClient != nil {
		limiter := ratelimit.New(redisClient, 10, 60*time.Second,
			ratelimit.WithKeyFunc(func(ip string) string { return fmt.Sprintf("ratelimit:onboard:%s", ip) }),
		)
		onboardHandler := middlewares.NewRateLimitMiddleware(limiter)(http.HandlerFunc(adminHandler.OnboardUser))
		router.HandleFunc(constants.OnboardUserEndpoint, onboardHandler.ServeHTTP).Methods("POST")
	} else {
		router.HandleFunc(constants.OnboardUserEndpoint, adminHandler.OnboardUser).Methods("POST")
	}

	router.HandleFunc(constants.CreatePollEndpoint, pollHandler.CreatePoll).Methods("POST")
	router.HandleFunc(constants.ClosePollEndpoint, pollHandler.ClosePoll).Methods("PATCH")
	router.HandleFunc(constants.GetPollsEndpoint, pollHandler.GetPolls).Methods("GET")
	router.HandleFunc(constants.UpdatePollEndpoint, pollHandler.UpdatePoll).Methods("PATCH")
	router.HandleFunc(constants.DeletePollEndpoint, pollHandler.DeletePoll).Methods("DELETE")
	router.HandleFunc(constants.PollByIdEndpoint, pollHandler.GetPollById).Methods("GET")
	router.HandleFunc(constants.CastVoteEndpoint, voteHandler.CastVote).Methods("POST")
	router.HandleFunc(constants.ReportEndpoint, pollHandler.GetPollReport).Methods("GET")

	port := fmt.Sprint(config.AppConfig.Port)
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: time.Duration(config.AppConfig.Timeout) * time.Second,
		ReadTimeout:  time.Duration(config.AppConfig.Timeout) * time.Second,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	if redisClient != nil {
		redisClient.Close()
	}

	postgres.Close(db)
	log.Info().Msg("Server exited")
}
