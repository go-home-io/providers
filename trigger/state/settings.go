//go:generate enumer -type=logic -transform=kebab -trimprefix=logic -json -text -yaml

package main

import (
	"errors"
	"sync"

	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/gobwas/glob"
)

// Describes possible logic.
type logic int

const (
	// Describes "OR" logic which is default.
	logicOr logic = iota
	// Describes "AND" logic.
	logicAnd
)

// DeviceEntry has data about single device this trigger has to watch.
type DeviceEntry struct {
	Device   string         `yaml:"device" validate:"required,gt=0"`
	Property enums.Property `yaml:"property" validate:"required"`
	State    interface{}    `yaml:"state"`
	Mapper   string         `yaml:"mapper"`

	mapperExpr   helpers.ITemplateExpression
	deviceRegexp glob.Glob
	triggered    bool
}

// Settings has data required to start trigger.
type Settings struct {
	sync.Mutex

	Delay       int            `yaml:"delay" validate:"gte=0"`
	Devices     []*DeviceEntry `yaml:"devices" validate:"required,gt=0"`
	Logic       string         `yaml:"logic" validate:"oneof=or and" default:"or"`
	Pessimistic bool           `yaml:"pessimistic" default:"-"`

	decisionLogic logic
	parser        helpers.ITemplateParser
}

// Validate validates supplied settings.
func (s *Settings) Validate() error {
	l, err := logicString(s.Logic)
	if err != nil {
		l = logicOr
	}

	s.parser = helpers.NewParser()
	s.decisionLogic = l

	for _, v := range s.Devices {
		if (nil == v.State && "" == v.Mapper) || (nil != v.State && "" != v.Mapper) {
			return errors.New("either state or mapper should be defined")
		}

		var err error
		v.deviceRegexp, err = glob.Compile(v.Device)
		if err != nil {
			return err
		}

		if "" != v.Mapper {
			exp, err := s.parser.Compile(v.Mapper)
			if err != nil {
				return err
			}

			v.mapperExpr = exp
			continue
		}

		v.State, err = helpers.PropertyFixYaml(v.State, v.Property)
		if err != nil {
			return err
		}
	}

	return nil
}
