package main

import "github.com/go-home-io/server/plugins/device/enums"

// Settings describes plugin settings.
type Settings struct {
	Name     string            `yaml:"name" validate:"required"`
	Prefix   string            `yaml:"topicsPrefix"`
	Devices  []*DeviceSettings `yaml:"devices" validate:"gt=0"`
	Login    string            `yaml:"login"`
	Password string            `yaml:"password"`
	ClientID string            `yaml:"clientID" validate:"required" default:"gohome"`
	Broker   string            `yaml:"broker" validate:"required"`
}

// Validate performs settings validation.
func (*Settings) Validate() error {
	return nil
}

// MQTTopicMapper defines data mappers from/to MQTT topics.
type MQTTopicMapper struct {
	Topic  string `yaml:"topic" validate:"required"`
	Mapper string `yaml:"mapper" validate:"required"`
}

// PropertyMapper defines properties mappers.
type PropertyMapper struct {
	MQTTopicMapper `yaml:",inline"`
	Property       enums.Property `yaml:"property" validate:"required"`
}

// CommandMapper defines commands mappers.
type CommandMapper struct {
	MQTTopicMapper `yaml:",inline"`
	Command        enums.Command `yaml:"command" validate:"required"`
}

// DeviceSettings defines single hub device mapper.
type DeviceSettings struct {
	Type        enums.DeviceType  `yaml:"type" validate:"required"`
	Name        string            `yaml:"name" validate:"required"`
	Qos         byte              `yaml:"qos" validate:"required,gt=0,lte=4" default:"2"`
	Retained    bool              `yaml:"retained" default:"false"`
	Pessimistic bool              `yaml:"pessimistic" default:"-"`
	Properties  []*PropertyMapper `yaml:"properties"`
	Commands    []*CommandMapper  `yaml:"commands"`
}
