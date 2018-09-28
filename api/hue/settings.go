package main

import (
	"math"

	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/gobwas/glob"
)

const (
	// Describes maximum possible brightness for the HUE lights.
	brightnessMax = math.MaxUint8 - 1
)

// Settings has data required to start API.
type Settings struct {
	AdvAddress     string             `yaml:"advAddress" validate:"required,ipv4port"`
	DeviceFilter   []string           `yaml:"devices"`
	DeviceTypes    []enums.DeviceType `yaml:"types"`
	NamesOverrides map[string]string  `yaml:"nameOverrides"`

	devRegexp []glob.Glob
	types     []enums.DeviceType
}

// Validate performs config validation.
func (s *Settings) Validate() error {
	if 0 == len(s.DeviceFilter) {
		s.DeviceFilter = []string{"**"}
	}

	s.types = make([]enums.DeviceType, 0)
	for _, v := range s.DeviceTypes {
		if !enums.SliceContainsDeviceType(supportedTypes, v) {
			continue
		}

		s.types = append(s.types, v)
	}

	if 0 == len(s.types) {
		s.types = supportedTypes
	}

	s.devRegexp = make([]glob.Glob, 0)
	for _, v := range s.DeviceFilter {
		a, err := glob.Compile(v)
		if err != nil {
			return err
		}

		s.devRegexp = append(s.devRegexp, a)
	}

	return nil
}

// List of supported device types.
var supportedTypes = []enums.DeviceType{enums.DevLight, enums.DevSwitch, enums.DevGroup, enums.DevVacuum}

// List of supported device properties.
var supportedProperties = []enums.Property{enums.PropOn, enums.PropBrightness}
