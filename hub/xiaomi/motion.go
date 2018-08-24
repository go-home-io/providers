package main // nolint: dupl

import (
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/vkorn/go-miio"
)

// Defines Xiaomi motion sensor.
type xiaomiMotion struct {
	xiaomiDevice
	state *device.SensorState
}

// GetSpec returns device spec.
func (m *xiaomiMotion) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropOn, enums.PropSensorType, enums.PropBatteryLevel},
		SupportedCommands:   []enums.Command{},
	}
}

// Load is not used since hub is responsible for device init.
func (m *xiaomiMotion) Load() (*device.SensorState, error) {
	return m.state, nil
}

// Update is not used since device pushes updates.
func (m *xiaomiMotion) Update() (*device.SensorState, error) {
	return nil, nil
}

// InternalUpdate performs internal state update in response of device messages.
func (m *xiaomiMotion) InternalUpdate(state interface{}, firstSeen bool) interface{} {
	s := state.(*miio.MotionState)
	m.state.On = s.HasMotion
	m.state.BatteryLevel = uint8(s.Battery)

	if !firstSeen {
		m.updatesChan <- &device.StateUpdateData{State: m.state}
	}

	return m.state
}
