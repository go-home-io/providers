package main

// Settings describes plugin settings.
type Settings struct {
	LookupAddress string `yaml:"lookup" validate:"required"`
	ServerAddress string `yaml:"server" validate:"required"`
	Timeout       int    `yaml:"timeout" validate:"gt=0" default:"1"`
}

// Validate settings.
func (s *Settings) Validate() error {
	return nil
}
