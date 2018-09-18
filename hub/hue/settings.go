//go:generate enumer -type=HueResources -transform=kebab -trimprefix=Resource

package main

import "time"

// HueResources describes enum with known hub resources.
type HueResources int

const (
	// ResourceAll describes all possible resources.
	ResourceAll HueResources = iota
	// ResourceLights describes lights resources.
	ResourceLights
	// ResourceGroups describes groups resources.
	ResourceGroups
)

// Settings describes plugin settings.
type Settings struct {
	BridgeIP        string   `yaml:"ip" validate:"isdefault|ipv4"`
	Token           string   `yaml:"token"`
	LoadResources   []string `yaml:"loadResources" validate:"unique,required,min=1,dive,oneof=* lights groups" default:"lights"` //nolint: lll
	PollingInterval int      `yaml:"pollingInterval" validate:"isdefault|numeric,gte=2" default:"20"`
	load            []HueResources
	pollingInterval time.Duration
}

// Validate settings.
func (s *Settings) Validate() error {
	allFound := false
	s.load = make([]HueResources, 0)
	for _, v := range s.LoadResources {
		resource, err := HueResourcesString(v)
		if err != nil {
			resource = ResourceAll
		}

		if ResourceAll == resource {
			allFound = true
			break
		} else {
			s.load = append(s.load, resource)
		}
	}

	if allFound {
		s.load = []HueResources{ResourceLights, ResourceGroups}
	}

	s.pollingInterval = time.Duration(s.PollingInterval) * time.Second
	return nil
}
