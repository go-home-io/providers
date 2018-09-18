// Package main contains Xiaomi implementation for the go-home vacuum.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}
	return &XiaomiVacuum{Settings: settings}, settings, nil
}
