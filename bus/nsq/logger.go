package main

import (
	"go-home.io/x/server/plugins/common"
)

// Override for nsq default logger.
type nsqLogger struct {
	logger common.ILoggerProvider
}

// Output overrides default nsq logger implementation.
func (l *nsqLogger) Output(depth int, s string) error {
	l.logger.Debug(s)
	return nil
}
