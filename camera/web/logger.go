package main

import (
	"github.com/bdlm/log"
	"github.com/go-home-io/server/plugins/common"
	"github.com/pkg/errors"
)

// Custom formatter implementation for the Chrome library.
type chromeLogger struct {
	Logger common.ILoggerProvider
	URL    string
}

// Format returns nothing to original logger,
// instead outputs to system logger.
func (l *chromeLogger) Format(e *log.Entry) ([]byte, error) {
	switch e.Level {
	case log.DebugLevel, log.InfoLevel, log.ErrorLevel:
		l.Logger.Debug(e.Message, common.LogURLToken, l.URL)
	case log.WarnLevel:
		l.Logger.Warn(e.Message, common.LogURLToken, l.URL)
	case log.FatalLevel, log.PanicLevel:
		l.Logger.Error(e.Message, errors.New("chrome error"), common.LogURLToken, l.URL)
	}

	return nil, nil
}
