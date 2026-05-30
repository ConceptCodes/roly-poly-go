package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"roly-poly/config"
	"roly-poly/internal/constants"
	_lg "roly-poly/pkg/logger"
)

func New() (*gorm.DB, error) {
	log := _lg.New()
	log.Debug().Msg("Connecting to postgres")

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.AppConfig.DbHost,
			config.AppConfig.DbUser,
			config.AppConfig.DbPass,
			config.AppConfig.DbName,
			config.AppConfig.DbPort,
			config.AppConfig.DbSslMode,
		),
	}), &gorm.Config{
		Logger: logger.New(
			log,
			logger.Config{
				LogLevel:             gormLogLevel(),
				Colorful:             config.AppConfig.Env == constants.LocalEnv,
				ParameterizedQueries: true,
			},
		),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	configurePool(sqlDB)

	return db, nil
}

func gormLogLevel() logger.LogLevel {
	switch config.AppConfig.Env {
	case constants.ProductionEnv:
		return logger.Silent
	default:
		return logger.Info
	}
}

func configurePool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)
}

func Close(db *gorm.DB) {
	log := _lg.New()
	sqlDB, err := db.DB()
	if err != nil {
		log.Error().Err(err).Msg("Error getting underlying sql.DB")
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing db")
	}
}

func HealthCheck(db *gorm.DB) bool {
	log := _lg.New()
	log.Debug().Msgf(constants.HealthCheckMessage, "postgres")

	sqlDB, err := db.DB()
	if err != nil {
		log.Error().Err(err).Msgf(constants.HealthCheckError, "postgres")
		return false
	}

	if err := sqlDB.Ping(); err != nil {
		log.Error().Err(err).Msgf(constants.HealthCheckError, "postgres")
		return false
	}

	log.Info().Msg("Postgres is up")
	return true
}
