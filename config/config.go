package config

import (
	"roly-poly/pkg/logger"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port    int `env:"PORT" envDefault:"8080"`
	Timeout int `env:"HTTP_TIMEOUT" envDefault:"15"`

	DbHost string `env:"DB_HOST" envDefault:"postgres_host"`
	DbPort int    `env:"DB_PORT" envDefault:"5432"`
	DbUser string `env:"DB_USER" envDefault:"postgres_user"`
	DbPass string `env:"DB_PASS" envDefault:"postgres_password"`
	DbName string `env:"DB_NAME" envDefault:"postgres_db"`

	Env string `env:"ENV" envDefault:"development"`
}

var AppConfig = Config{}

func init() {
	log := logger.New()
	log.Debug().Msg("Loading env vars")

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error while loading env vars")
	}

	err = env.Parse(&AppConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while parsing env vars")
	}
}
