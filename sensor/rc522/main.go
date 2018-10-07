// +build linux,arm

// Package main contains rc522 implementation for the go-home presence sensor.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &RC522Sensor{Settings: settings}, settings, nil
}
