package main

import "time"

// Settings has data required to start sensor.
type Settings struct {
	DeviceID        int `yaml:"device" default:"1"`
	Address         int `yaml:"address" default:"0x76"`
	PollingInterval int `yaml:"pollingInterval" validate:"gt=2" default:"30"`

	updateInterval time.Duration
}

// Validate checks settings.
func (s *Settings) Validate() error {
	s.updateInterval = time.Duration(s.PollingInterval) * time.Second
	return nil
}
