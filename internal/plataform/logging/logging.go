package logging

import (
	"strings"

	"github.com/duddy57/sperium/internal/plataform/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(service string) (*zap.Logger, error) {
	var l *zap.Logger
	var err error
	switch strings.ToLower(config.String("APP_ENV", "development")) {
	case "production":
		l, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	case "development":
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		l, err = cfg.Build()
		if err != nil {
			return nil, err
		}
	}

	l = l.Named(service)
	return l, nil
}

func levelFromEnv() zapcore.Level {
	switch strings.ToLower(config.String("LOG_LEVEL", "INFO")) {
	case "error":
		return zap.ErrorLevel
	case "warn":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}
