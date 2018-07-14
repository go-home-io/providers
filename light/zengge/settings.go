package main

import "time"

// Settings describes plugin settings.
type Settings struct {
	LightIP         string `yaml:"ip" validate:"ipv4"`
	PollingInterval int    `yaml:"pollingInterval" validate:"numeric,gte=2" default:"5"`

	pollingInterval time.Duration
}

const (
	// Describes default Zengge port
	devicePort = 5577
)

// Validate settings.
func (s *Settings) Validate() error {
	s.pollingInterval = time.Duration(s.PollingInterval) * time.Second
	return nil
}
