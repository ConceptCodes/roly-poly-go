package postgres

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"roly-poly/config"
	"roly-poly/internal/constants"
	_lg "roly-poly/pkg/logger"
)

var (
	db   *gorm.DB
	once sync.Once
	log  *zerolog.Logger
)

func init() {
	log = _lg.New()
}

func New() (*gorm.DB, error) {
	var err error
	log.Debug().Msg("Connecting to postgres")

	once.Do(func() {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
				config.AppConfig.DbHost,
				config.AppConfig.DbUser,
				config.AppConfig.DbPass,
				config.AppConfig.DbName,
				config.AppConfig.DbPort,
			),
		}), &gorm.Config{
			Logger: logger.New(
				log,
				logger.Config{
					LogLevel:             logger.Info,
					Colorful:             config.AppConfig.Env == constants.LocalEnv,
					ParameterizedQueries: true,
				},
			),
		})

	})
	return db, err
}

func Close() {
	sqlDB, _ := db.DB()

	err := sqlDB.Close()

	if err != nil {
		log.Error().Err(err).Msg("Error closing db")
		return
	}
}

func HealthCheck() bool {
	log := _lg.New()
	log.Debug().Msgf(constants.HealthCheckMessage, "postgres")

	sqlDB, _ := db.DB()

	err := sqlDB.Ping()

	if err != nil {
		log.
			Error().
			Err(err).
			Msgf(constants.HealthCheckError, "postgres")
		return false
	}

	log.Info().Msg("Postgres is up")
	return true
}
