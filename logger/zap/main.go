// Package main contains ZAP implementation for the go-home logger.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &ZapLogger{Settings: settings}, settings, nil
}
