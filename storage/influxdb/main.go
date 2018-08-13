// Package main contains InfluxDB implementation of the go-home state storage.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &InfluxStorage{Settings: settings}, settings, nil
}
