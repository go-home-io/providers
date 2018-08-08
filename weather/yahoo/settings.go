package main

import (
	"strings"
	"time"

	"github.com/go-home-io/server/plugins/device/enums"
)

// Settings has data required to start weather.
type Settings struct {
	Location        string           `yaml:"location" validate:"required"`
	Properties      []enums.Property `yaml:"properties"`
	PollingInterval int              `yaml:"pollingInterval" validate:"gte=10" default:"10"`

	updateInterval time.Duration
}

// Validate validates supplied settings.
func (s *Settings) Validate() error {
	if 0 == len(s.Properties) {
		s.Properties = supportedProperties
	} else {
		final := make([]enums.Property, 0)
		for _, v := range s.Properties {
			if enums.SliceContainsProperty(supportedProperties, v) {
				final = append(final, v)
			}
		}

		s.Properties = final
	}

	s.updateInterval = time.Duration(s.PollingInterval) * time.Minute

	// YQL doesn't like spaces
	s.Location = strings.Replace(s.Location, " ", "", -1)
	return nil
}

var supportedProperties = []enums.Property{enums.PropTemperature, enums.PropSunrise, enums.PropSunset,
	enums.PropHumidity, enums.PropPressure, enums.PropVisibility, enums.PropWindDirection, enums.PropWindSpeed}
