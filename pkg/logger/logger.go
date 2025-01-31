package logger

import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "os"
)

func InitLogger() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func Info(message string) {
    log.Info().Msg(message)
}

func Error(message string) {
    log.Error().Msg(message)
}
