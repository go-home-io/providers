// Package main contains telegram implementation for the go-home notification.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}
	return &TelegramNotification{Settings: settings}, settings, nil
}
