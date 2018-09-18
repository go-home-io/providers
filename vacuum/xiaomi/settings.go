package main

// Settings defines plugin settings.
type Settings struct {
	IP  string `yaml:"ip" validate:"required,ipv4"`
	Key string `yaml:"key" validate:"required"`
}

// Validate is not used.
func (s *Settings) Validate() error {
	return nil
}
