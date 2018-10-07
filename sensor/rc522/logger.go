package main

import "github.com/sirupsen/logrus"

// Fake logger for the library
type fakeLogger struct {
}

// Format prevents data to be printed.
func (l *fakeLogger) Format(e *logrus.Entry) ([]byte, error) {
	return nil, nil
}
