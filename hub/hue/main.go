// Package main contains Philips HUE implementation for the go-home hub.
package main

import "math"

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &HueHub{
		sharedObjects: &sharedObjects{
			settings:             settings,
			internalLightUpdates: make(chan int, 5),
		},
	}, settings, nil
}

const (
	// Describes maximum possible brightness for the HUE lights.
	brightnessMax = math.MaxUint8 - 1
	// Describes number of iterations per second when gradually adjusting brightness.
	overtimeBrightnessStepsPerSecond = 2
)
