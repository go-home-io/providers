package main // nolint: dupl

import (
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/vkorn/go-miio"
)

// Defines Xiaomi magnet.
type xiaomiMagnet struct {
	xiaomiDevice
	state *device.SensorState
}

// GetSpec returns device spec.
func (m *xiaomiMagnet) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropOn, enums.PropSensorType, enums.PropBatteryLevel},
		SupportedCommands:   []enums.Command{},
	}
}

// Load is not used since hub is responsible for device init.
func (m *xiaomiMagnet) Load() (*device.SensorState, error) {
	return m.state, nil
}

// Update is not used since device pushes updates.
func (m *xiaomiMagnet) Update() (*device.SensorState, error) {
	return nil, nil
}

// InternalUpdate performs internal state update in response of device messages.
func (m *xiaomiMagnet) InternalUpdate(state interface{}, firstSeen bool) interface{} {
	s := state.(*miio.MagnetState)
	m.state.On = s.Opened
	m.state.BatteryLevel = uint8(s.Battery)

	if !firstSeen {
		m.updatesChan <- &device.StateUpdateData{State: m.state}
	}

	return m.state
}
