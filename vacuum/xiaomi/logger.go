package main

import (
	"errors"
	"fmt"

	"github.com/go-home-io/server/plugins/common"
)

// Implements internal logger.
type xiaomiLogger struct {
	logger common.ILoggerProvider
}

// Debug does nothing since library logs are too extensive.
func (l *xiaomiLogger) Debug(format string, v ...interface{}) {
}

// Info logs debug level since library logs are too extensive.
func (l *xiaomiLogger) Info(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

// Warn logs info level since library logs are too extensive.
func (l *xiaomiLogger) Warn(format string, v ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, v...))
}

// Error logs error level.
func (l *xiaomiLogger) Error(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...), errors.New("xiaomi error"))
}

// Fatal logs error level.
func (l *xiaomiLogger) Fatal(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...), errors.New("xiaomi error"))
}
