package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(
	preset string,
	level string,
	encoding string,
	outputPaths []string,
	errOutputPaths []string,
) (*zap.Logger, error) {
	config := zap.Config{}

	if preset == "dev" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	if level != "" {
		logLevel, err := zapcore.ParseLevel(level)
		if err != nil {
			return nil, fmt.Errorf("parse log level: %w", err)
		}
		config.Level = zap.NewAtomicLevelAt(logLevel)
	}

	if encoding != "" {
		config.Encoding = encoding
	}
	if config.Encoding == "console" {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if len(outputPaths) > 0 {
		config.OutputPaths = outputPaths
	}
	if len(errOutputPaths) > 0 {
		config.ErrorOutputPaths = errOutputPaths
	}

	l, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	return l, nil
}
