package lib

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

func InitLogger(logLevel zapcore.Level) *zap.Logger {

	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 format for human-readable time
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Encoding = "console"

	var err error
	logger, err = config.Build()
	if err != nil {
		panic("Failed to initialize logger")
	}

	return logger
}
