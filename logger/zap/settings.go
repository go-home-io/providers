//go:generate enumer -type=target -json -text -yaml

package main

import "github.com/pkg/errors"

type target int

const (
	// Regular console output.
	console target = iota
	// InfluxDB core.
	influxDB
)

// InfluxSettings describes influx DB settings.
type InfluxSettings struct {
	Address   string `yaml:"address" validate:"required"`
	Username  string `yaml:"username" validate:"required"`
	Password  string `yaml:"password" validate:"required"`
	Database  string `yaml:"database" validate:"required"`
	BatchSize int    `yaml:"batchSize" validate:"gt=0" default:"10"`
}

// Settings describes plugin settings.
type Settings struct {
	Target string          `yaml:"target" validate:"required,oneof=console influxDB" default:"console"`
	Influx *InfluxSettings `yaml:"influxDB"`

	targetCore target
}

// Validate settings.
func (s *Settings) Validate() error {
	t, err := targetString(s.Target)
	if err != nil {
		return err
	}

	if t == influxDB && nil == s.Influx {
		return errors.New("influxDB connection details are empty")
	}

	s.targetCore = t
	return nil
}
