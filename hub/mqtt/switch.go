package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// MQTTSwitch implements switch interface.
type MQTTSwitch struct {
	mqttDevice
}

// Constructs a new switch.
// nolint:dupl
func newSwitch(topicPrefix string, parser helpers.ITemplateParser, settings *DeviceSettings, client mqtt.Client,
	logger common.ILoggerProvider, uom enums.UOM) *MQTTSwitch {
	s := &MQTTSwitch{
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

// Load performs initial sensor load.
func (m *MQTTSwitch) Load() (*device.SwitchState, error) {
	m.state = &device.SwitchState{
		On: false,
	}

	m.subscribe()

	if enums.SliceContainsCommand(m.spec.SupportedCommands, enums.CmdOn) &&
		enums.SliceContainsCommand(m.spec.SupportedCommands, enums.CmdOff) {
		m.spec.SupportedCommands = append(m.spec.SupportedCommands, enums.CmdToggle)
	}

	return m.state.(*device.SwitchState), nil
}

// On sends ON command to corresponding MQTT topic.
func (m *MQTTSwitch) On() error {
	err := m.command(enums.CmdOn)
	if err != nil {
		return errors.Wrap(err, "on command failed")
	}

	if m.settings.Pessimistic {
		m.state.(*device.SwitchState).On = true
		m.forceUpdate()
	}

	return nil
}

// Off sends OFF command to corresponding MQTT topic.
func (m *MQTTSwitch) Off() error {
	err := m.command(enums.CmdOff)
	if err != nil {
		return errors.Wrap(err, "off command failed")
	}

	if !m.settings.Pessimistic {
		m.state.(*device.SwitchState).On = true
		m.forceUpdate()
	}

	return nil
}

// Toggle sends ON/OFF command to corresponding MQTT topic.
func (m *MQTTSwitch) Toggle() error {
	if m.state.(*device.SwitchState).On {
		return m.Off()
	}

	return m.On()
}

// Update is unused for MQTT hub because we're not polling it.
// Instead we're using callback chan.
func (m *MQTTSwitch) Update() (*device.SwitchState, error) {
	return nil, nil
}
