// Package main contains yahoo implementation for the go-home weather.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &YahooWeather{Settings: settings}, settings, nil
}
