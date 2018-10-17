// Package main contains MQTT implementation for the go-home hub.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &MQTTHub{
		Settings: settings,
	}, settings, nil
}

const (
	// Log topic representation.
	logTokenTopic = "topic"
	// Log expected device type representation.
	logTokenExpectedType = "expected_type"
)
