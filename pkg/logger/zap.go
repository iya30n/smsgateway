package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

const (
	levelEnv = "LOG_LEVEL"
	dirEnv   = "LOG_DIR"
)

func init() {
	log, sinkErr := newLogger()
	Logger = log

	if sinkErr != nil {
		Logger.Warn("file log sink disabled", zap.Error(sinkErr))
	}
}

func newLogger() (*zap.Logger, error) {
	level := parseLevel(os.Getenv(levelEnv))

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	cores := []zapcore.Core{zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)}

	writer, err := newFileWriter(os.Getenv(dirEnv))
	if err != nil {
		return zap.New(zapcore.NewTee(cores...),
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		), err
	}

	cores = append(cores, zapcore.NewCore(encoder, writer, level))

	return zap.New(zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	), nil
}

func newFileWriter(dir string) (zapcore.WriteSyncer, error) {
	if dir == "" {
		dir = "./logs"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return zapcore.AddSync(&lumberjack.Logger{
		Filename:  dir + "/log.json",
		LocalTime: false,
		MaxSize:   10, // megabytes
		MaxAge:    30, // days
	}), nil
}

func parseLevel(value string) zapcore.Level {
	switch value {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "info", "":
		return zapcore.InfoLevel
	default:
		return zapcore.InfoLevel
	}
}
