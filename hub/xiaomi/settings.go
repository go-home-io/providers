package main

import "go-home.io/x/server/plugins/device/enums"

// Settings defines plugin settings.
type Settings struct {
	IP  string    `yaml:"ip" validate:"required,ipv4"`
	Key string    `yaml:"key" validate:"required"`
	UOM enums.UOM `yaml:"units" default:"metric"`
}

// Validate is not used.
func (s *Settings) Validate() error {
	return nil
}
