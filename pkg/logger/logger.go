package logger

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"

	"roly-poly/config"
	"roly-poly/internal/constants"
)

var (
	instance *zerolog.Logger
	once     sync.Once
)

func New() *zerolog.Logger {
	once.Do(func() {
		env := config.AppConfig.Env
		var output io.Writer = os.Stdout

		if env == constants.LocalEnv || env == constants.DevelopmentEnv {
			output = zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: constants.TimeFormat,
			}
		}

		log := zerolog.New(output).
			With().
			Timestamp().
			CallerWithSkipFrameCount(2).
			Logger()

		instance = &log
	})
	return instance
}
