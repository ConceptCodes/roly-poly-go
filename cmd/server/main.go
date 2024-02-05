package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"roly-poly/config"
	"roly-poly/internal/constants"
	"roly-poly/internal/handlers"
	"roly-poly/internal/middlewares"
	"roly-poly/internal/models"
	repository "roly-poly/internal/repositories"
	"roly-poly/pkg/logger"
	"roly-poly/pkg/storage/postgres"
)

var (
	db  *gorm.DB
	err error
)

func Run() {
	db, err = postgres.New()
	log := logger.New()

	if err != nil {
		log.Fatal().Err(err).Msg("Error while connecting to database")
	}

	db.AutoMigrate(
		&models.UserModel{},
		&models.PollModel{},
		&models.OptionModel{},
		&models.VoteModel{},
	)

	userRepo := repository.NewGormUserRepository(db)
	pollRepo := repository.NewGormPollRepository(db)
	optionRepo := repository.NewGormOptionRepository(db)

	healthHandler := handlers.NewHealthHandler()
	adminHandler := handlers.NewAdminHandler(userRepo)
	pollHandler := handlers.NewPollHandler(pollRepo, optionRepo)

	router := mux.NewRouter()

	// Middlewares
	router.Use(middlewares.TraceRequest)
	router.Use(middlewares.ContentTypeJSON)
	router.Use(middlewares.AuthMiddleware)
	router.Use(middlewares.RequestLogger)

	// Routes
	router.HandleFunc(constants.HealthCheckEndpoint, healthHandler.ServiceAliveHandler).Methods("GET")
	router.HandleFunc(constants.ReadinessEndpoint, healthHandler.ServiceReadyHandler).Methods("GET")

	router.HandleFunc(constants.OnboardUserEndpoint, adminHandler.OnboardUser).Methods("POST")

	router.HandleFunc(constants.CreatePollEndpoint, pollHandler.CreatePoll).Methods("POST")
	router.HandleFunc(constants.ClosePollEndpoint, pollHandler.ClosePoll).Methods("GET")
	router.HandleFunc(constants.GetPollsEndpoint, pollHandler.GetPolls).Methods("GET")
	router.HandleFunc(constants.UpdatePollEndpoint, pollHandler.UpdatePoll).Methods("PATCH")
	router.HandleFunc(constants.PollByIdEndpoint, pollHandler.GetPollById).Methods("GET")

	port := fmt.Sprint(config.AppConfig.Port)
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("127.0.0.1:%s", port),
		WriteTimeout: time.Duration(config.AppConfig.Timeout) * time.Second,
		ReadTimeout:  time.Duration(config.AppConfig.Timeout) * time.Second,
	}

	log.Debug().Msgf(constants.StartMessage, port, config.AppConfig.Env)

	err = srv.ListenAndServe()

	if err != nil {
		postgres.Close()
		log.
			Fatal().
			Err(err).
			Msg("Error while starting server")
	}
}
