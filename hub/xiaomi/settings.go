package main

import "github.com/go-home-io/server/plugins/device/enums"

// Settings defines plugin settings.
type Settings struct {
	IP  string    `yaml:"ip" validate:"required,ipv4"`
	Key string    `yaml:"key" validate:"required"`
	UOM enums.UOM `yaml:"uom" default:"imperial"`
}

// Validate is not used.
func (s *Settings) Validate() error {
	return nil
}
