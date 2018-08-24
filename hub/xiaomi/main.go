// Package main contains Xiaomi implementation for the go-home hub.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &XiaomiHub{
		Settings: settings,
	}, settings, nil
}
