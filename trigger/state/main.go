// Package main contains devices' state update implementation for the go-home trigger.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}
	return &StateTrigger{Settings: settings}, settings, nil
}
