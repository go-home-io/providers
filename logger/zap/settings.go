package main

// Settings describes plugin settings.
type Settings struct {
	Targets struct {
		Regular []string `yaml:"regular" validate:"unique,min=1"`
		Error   []string `yaml:"error" validate:"unique,min=1"`
	} `yaml:"targets" validate:"required"`
}

// Validate settings.
func (s *Settings) Validate() error {
	return nil
}
