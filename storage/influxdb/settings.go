package main

import "time"

// Settings contains data required for instantiating a new influx provider.
type Settings struct {
	Address   string `yaml:"address" validate:"required"`
	Username  string `yaml:"username" validate:"required"`
	Password  string `yaml:"password" validate:"required"`
	Database  string `yaml:"database" validate:"required"`
	BatchSize int    `yaml:"batchSize" validate:"gt=0" default:"10"`
	Retention string `yaml:"retention" default:"7d"`
}

// Validate is not used for this plugin.
func (s *Settings) Validate() error {
	return nil
}

// Returns current UTC time.
func now() time.Time {
	return time.Now().UTC()
}

const (
	// Event measurement.
	eventToken = "event"
	// Event state.
	stateToken = "state"
	// Event ping.
	heartbeatToken = "ping"
)
