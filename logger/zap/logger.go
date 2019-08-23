package main

import (
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger describes ZAP logger implementation.
type ZapLogger struct {
	logger     *zap.Logger
	Settings   *Settings
	hasHistory bool
	core       IHistoryCore
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
		Encoding: "console",
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
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	options := []zap.Option{
		zap.AddCallerSkip(data.SkipLevel),
		zap.AddStacktrace(zap.ErrorLevel),
	}

	switch l.Settings.targetCore {
	case influxDB:
		l.hasHistory = true
		core := newInfluxCore(logLevel, l.Settings.Influx)
		err := core.CreateDatabase()
		if err != nil {
			return errors.Wrap(err, "failed to provision logs database")
		}
		options = append(options, zap.WrapCore(func(zapcore.Core) zapcore.Core {
			return core
		}))
		l.core = core
	case console:
		l.hasHistory = false
	}

	z, err := zConfig.Build(options...)
	if err != nil {
		return errors.Wrap(err, "zap build failed")
	}

	l.logger = z
	return nil
}

// GetSpecs returns logger specifications.
func (l *ZapLogger) GetSpecs() *common.LogSpecs {
	return &common.LogSpecs{
		IsHistorySupported: l.hasHistory,
	}
}

// Query requests log history.
func (l *ZapLogger) Query(r *common.LogHistoryRequest) []*common.LogHistoryEntry {
	if !l.hasHistory {
		return nil
	}

	return l.core.Query(r)
}

// Debug outputs debug level message.
func (l *ZapLogger) Debug(msg string, fields ...string) {
	params := convertParams(fields...)
	l.logger.Debug(msg, params...)
}

// Info outputs info level message.
func (l *ZapLogger) Info(msg string, fields ...string) {
	params := convertParams(fields...)
	l.logger.Info(msg, params...)
}

// Warn outputs warning level message.
func (l *ZapLogger) Warn(msg string, fields ...string) {
	params := convertParams(fields...)
	l.logger.Warn(msg, params...)
}

// Error outputs error level message.
func (l *ZapLogger) Error(msg string, fields ...string) {
	params := convertParams(fields...)
	l.logger.Error(msg, params...)
}

// Fatal outputs fatal level message end performs os.Exit(1).
func (l *ZapLogger) Fatal(msg string, fields ...string) {
	params := convertParams(fields...)
	defer l.logger.Sync() // nolint: errcheck
	l.logger.Fatal(msg, params...)
}

// Converts input string params into zap.Fields.
func convertParams(fields ...string) []zap.Field {
	fLen := len(fields)
	result := make([]zap.Field, fLen/2)
	for ii := 0; ii < fLen; ii += 2 {
		if ii+1 >= fLen {
			break
		}

		result[ii/2] = zap.String(fields[ii], fields[ii+1])
	}

	return result
}
