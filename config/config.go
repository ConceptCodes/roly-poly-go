package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port           int `env:"PORT" envDefault:"8080"`
	ReadTimeout    int `env:"HTTP_READ_TIMEOUT" envDefault:"15"`
	WriteTimeout   int `env:"HTTP_WRITE_TIMEOUT" envDefault:"30"`
	ShutdownTimeout int `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10"`

	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"DB_USER" envDefault:"postgres"`
	DbPass     string `env:"DB_PASS" envDefault:"postgres"`
	DbName     string `env:"DB_NAME" envDefault:"roly_poly"`
	DbSslMode  string `env:"DB_SSL_MODE" envDefault:"require"`

	RedisHost string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPass string `env:"REDIS_PASS" envDefault:""`

	CorsAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	OtelEndpoint   string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	OtelServiceName string `env:"OTEL_SERVICE_NAME" envDefault:"roly-poly"`

	Env string `env:"ENV" envDefault:"development"`
}

var AppConfig = Config{}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("INFO: no .env file found, using environment variables")
	}

	if err := env.Parse(&AppConfig); err != nil {
		log.Fatalf("FATAL: error parsing env vars: %v", err)
	}
}
