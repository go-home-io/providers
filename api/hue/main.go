// Package main contains Philips HUE emulator implementation for the go-home extended API.
// Implementation is based on https://github.com/mdempsky/huejack
// More info about discovery http://www.burgestrand.se/hue-api/api/discovery/
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &HueEmulator{
			Settings:           settings,
			devices:            make(map[string]*DeviceUpdateMessage),
			unsupportedDevices: make([]string, 0),
			chCommands:         make(chan []byte, 5),
		},
		settings, nil
}
