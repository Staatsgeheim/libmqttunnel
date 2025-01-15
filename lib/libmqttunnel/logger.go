package main

import (
	"C"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func setupLog(logLevel C.int) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig = newEncoderConfig()
	
	if C.int(logLevel) == 0 {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else if C.int(logLevel) == 1 {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else if C.int(logLevel) == 2 {
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	} else if C.int(logLevel) == 3 {
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	} else if C.int(logLevel) == 6 {
		cfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}
	cfg.Encoding = "json"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	zapLogger, _ := cfg.Build()
	logger = zapLogger.Sugar()
	return zapLogger
}

func newEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.LevelKey = "level"
	cfg.MessageKey = "message"
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	return cfg
}
