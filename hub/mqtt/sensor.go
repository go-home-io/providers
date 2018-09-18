package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
)

// MQTTSensor implements sensor interface.
type MQTTSensor struct {
	mqttDevice
}

// Constructs a new sensor.
// nolint:dupl
func newSensor(topicPrefix string, parser helpers.ITemplateParser, settings *DeviceSettings, client mqtt.Client,
	logger common.ILoggerProvider, uom enums.UOM) *MQTTSensor {
	s := &MQTTSensor{
		mqttDevice: mqttDevice{
			settings:     settings,
			client:       client,
			parser:       parser,
			topicsPrefix: topicPrefix,
			logger:       logger,
			uom:          uom,
		},
	}

	return s
}

// Update is unused for MQTT hub because we're not polling it.
// Instead we're using callback chan.
func (m *MQTTSensor) Update() (*device.SensorState, error) {
	return nil, nil
}

// Load performs initial sensor load.
func (m *MQTTSensor) Load() (*device.SensorState, error) {
	m.state = &device.SensorState{}
	m.subscribe()
	return m.state.(*device.SensorState), nil
}
