package main

import "github.com/gorhill/cronexpr"

// Settings has data required to start trigger.
type Settings struct {
	Schedule string `yaml:"schedule" validate:"required"`

	expr *cronexpr.Expression
}

// Validate validates supplied settings.
func (s *Settings) Validate() error {
	expr, err := cronexpr.Parse(s.Schedule)
	if err != nil {
		return err
	}

	s.expr = expr
	return nil
}
