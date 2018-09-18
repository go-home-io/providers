package main

import (
	"sync"

	"github.com/go-home-io/server/plugins/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger describes ZAP logger implementation.
type ZapLogger struct {
	sync.Mutex
	logger   *zap.Logger
	Settings *Settings
}

// Init performs initial logger setup.
func (l *ZapLogger) Init(data *logger.InitDataLogger) error {
	var logLevel zapcore.Level
	switch data.Level {
	case logger.Debug:
		logLevel = zap.DebugLevel
	case logger.Error:
		logLevel = zap.ErrorLevel
	case logger.Warning:
		logLevel = zap.WarnLevel
	default:
		logLevel = zap.InfoLevel
	}

	zConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      l.Settings.Targets.Regular,
		ErrorOutputPaths: l.Settings.Targets.Error,
	}

	z, err := zConfig.Build(zap.AddCallerSkip(2),
		zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return err
	}
	l.logger = z
	return nil
}

// Debug outputs debug level message.
func (l *ZapLogger) Debug(msg string, fields ...string) {
	params := convertParams(fields...)
	l.Lock()
	defer l.Unlock()
	l.logger.Debug(msg, params...)
}

// Info outputs info level message.
func (l *ZapLogger) Info(msg string, fields ...string) {
	params := convertParams(fields...)
	l.Lock()
	defer l.Unlock()
	l.logger.Info(msg, params...)
}

// Warn outputs warning level message.
func (l *ZapLogger) Warn(msg string, fields ...string) {
	params := convertParams(fields...)
	l.Lock()
	defer l.Unlock()
	l.logger.Warn(msg, params...)
}

// Error outputs error level message.
func (l *ZapLogger) Error(msg string, fields ...string) {
	params := convertParams(fields...)
	l.Lock()
	defer l.Unlock()
	l.logger.Error(msg, params...)
}

// Fatal outputs fatal level message end performs os.Exit(1).
func (l *ZapLogger) Fatal(msg string, fields ...string) {
	params := convertParams(fields...)
	l.Lock()
	defer l.Unlock()
	defer l.logger.Sync()
	l.logger.Fatal(msg, params...)
}

// Flush performs logger buffer flush if any.
func (l *ZapLogger) Flush() {
	l.Lock()
	defer l.Unlock()
	l.logger.Sync() // nolint: gosec
}

// Converts input string params into zap.Fields.
func convertParams(fields ...string) []zap.Field {
	fLen := len(fields)
	result := make([]zap.Field, int(fLen/2))
	for ii := 0; ii < fLen; ii += 2 {
		if ii+1 >= fLen {
			break
		}

		result[int(ii/2)] = zap.String(fields[ii], fields[ii+1])
	}

	return result
}
