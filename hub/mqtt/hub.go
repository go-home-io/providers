package main

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// MQTTHub implements hub interface.
type MQTTHub struct {
	Settings   *Settings
	client     mqtt.Client
	updateChan chan *device.StateUpdateData
	parser     helpers.ITemplateParser
	logger     common.ILoggerProvider
	spec       *device.Spec
	uom        enums.UOM

	devices []interface{}
}

// Init performs initial plugin initialization.
func (m *MQTTHub) Init(data *device.InitDataDevice) error {
	m.uom = data.UOM
	m.updateChan = data.DeviceStateUpdateChan
	m.parser = helpers.NewParser()
	m.logger = data.Logger
	m.spec = &device.Spec{
		UpdatePeriod:        0,
		SupportedProperties: []enums.Property{enums.PropNumDevices},
		SupportedCommands:   []enums.Command{},
	}

	options := mqtt.NewClientOptions().
		SetClientID(m.Settings.ClientID).
		AddBroker(fmt.Sprintf("tcp://%s", m.Settings.Broker)).
		SetUsername(m.Settings.Login).
		SetPassword(m.Settings.Password).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(1 * time.Second).
		SetPingTimeout(1 * time.Second)

	firstConnect := true

	options.OnConnect = func(client mqtt.Client) {
		if firstConnect {
			firstConnect = false
			return
		}

		for _, v := range m.devices {
			v.(IGenericDevice).ReConnect()
			time.Sleep(50 * time.Millisecond)
		}
	}

	m.client = mqtt.NewClient(options)
	return nil
}

// Unload calls for broker disconnect.
func (m *MQTTHub) Unload() {
	m.client.Disconnect(2000)
}

// GetName returns device name.
func (m *MQTTHub) GetName() string {
	parts := strings.Split(m.Settings.Broker, ":")
	return parts[0]
}

// GetSpec returns device spec.
func (m *MQTTHub) GetSpec() *device.Spec {
	return m.spec
}

// Load performs initial hub load.
func (m *MQTTHub) Load() (*device.HubLoadResult, error) {
	m.devices = make([]interface{}, 0)
	token := m.client.Connect()
	token.WaitTimeout(2 * time.Second)
	if !m.client.IsConnected() {
		return nil, errors.New("connection to broker failed")
	}

	result := &device.HubLoadResult{
		Devices: make([]*device.DiscoveredDevices, 0),
	}

	for _, v := range m.Settings.Devices {
		var s device.IDevice
		var state interface{}
		var err error

		switch v.Type {
		case enums.DevSwitch:
			s = newSwitch(m.Settings.Prefix, m.parser, v, m.client, m.logger, m.uom)
			state, err = s.(*MQTTSwitch).Load()
		case enums.DevSensor:
			s = newSensor(m.Settings.Prefix, m.parser, v, m.client, m.logger, m.uom)
			state, err = s.(*MQTTSensor).Load()
		default:
			m.logger.Warn("This MQTT device type is unsupported", logTokenExpectedType, v.Type.String())
			continue
		}

		if err != nil {
			m.logger.Error("Failed to initialize MQTT device", err,
				"name", v.Name, logTokenExpectedType, v.Type.String())
			continue
		}

		m.devices = append(m.devices, s)
		disc := &device.DiscoveredDevices{
			Type:      v.Type,
			State:     state,
			Interface: s,
		}
		result.Devices = append(result.Devices, disc)
	}

	result.State = &device.HubState{
		NumDevices: len(m.devices),
	}

	m.logger.Info("Successfully connected to MQTT broker", common.LogDeviceHostToken, m.Settings.Broker)
	return result, nil
}

// Update is unused for MQTT hub because we're not polling it.
// Instead we're using callback chan.
func (m *MQTTHub) Update() (*device.HubLoadResult, error) {
	return nil, nil
}
