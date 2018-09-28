package main

import "strings"

// Settings describes device settings.
type Settings struct {
	Address         string `yaml:"address" validate:"required"`
	ChromeAddress   string `yaml:"chromeAddress" validate:"required"`
	ChromePort      int    `yaml:"chromePort" validate:"required" default:"9222"`
	PollingInterval int    `yaml:"pollingInterval" validate:"gt=10" default:"30"`
	ReloadInterval  int    `yaml:"reloadInterval" default:"0"`
	Width           int    `yaml:"width" default:"800"`
	Height          int    `yaml:"height" default:"600"`
}

// Validate performs settings validation.
func (s *Settings) Validate() error {
	if !strings.HasPrefix(s.Address, "http") {
		s.Address = "http://" + s.Address
	}

	return nil
}
