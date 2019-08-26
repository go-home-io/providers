// Package main contains August smart lock implementation for the go-home lock.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &AugustLock{Settings: settings}, settings, nil
}
