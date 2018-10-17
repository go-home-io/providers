// Package main contains cron implementation for the go-home trigger.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}
	return &CronTrigger{Settings: settings}, settings, nil
}
