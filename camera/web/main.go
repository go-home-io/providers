// Package main contains web implementation for the go-home camera.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &WebCamera{Settings: settings}, settings, nil
}
