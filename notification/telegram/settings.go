package main

// Settings has data required to start notification system.
type Settings struct {
	Token  string `yaml:"token" validate:"required"`
	ChatID string `yaml:"chat" validate:"required"`
}

// Validate validates supplied settings.
func (s *Settings) Validate() error {
	return nil
}
