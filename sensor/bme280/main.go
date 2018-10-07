// +build linux,arm

// Package main contains bme280 implementation for the go-home temperature sensor.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &BME280Sensor{Settings: settings}, settings, nil
}
