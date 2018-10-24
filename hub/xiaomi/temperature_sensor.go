package main

import (
	"github.com/vkorn/go-miio"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// Defines Xiaomi humidity/temperature sensor.
type xiaomiTemperatureSensor struct {
	xiaomiDevice
	state      *device.SensorState
	desiredUOM enums.UOM
	currentUOM enums.UOM
}

// Init performs initial device init. We need to overwrite it since
// UOM is required for this device.
func (s *xiaomiTemperatureSensor) Init(data *device.InitDataDevice) error {
	s.updatesChan = data.DeviceStateUpdateChan
	s.desiredUOM = data.UOM
	return nil
}

// GetSpec returns device spec.
func (s *xiaomiTemperatureSensor) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropTemperature, enums.PropHumidity,
			enums.PropSensorType, enums.PropBatteryLevel},
		SupportedCommands: []enums.Command{},
	}
}

// Load is not used since hub is responsible for device init.
func (s *xiaomiTemperatureSensor) Load() (*device.SensorState, error) {
	return s.state, nil
}

// Update is not used since device pushes updates.
func (s *xiaomiTemperatureSensor) Update() (*device.SensorState, error) {
	return nil, nil
}

// InternalUpdate performs internal state update in response of device messages.
func (s *xiaomiTemperatureSensor) InternalUpdate(state interface{}, firstSeen bool) interface{} {
	st := state.(*miio.SensorHTState)
	s.state.Temperature = helpers.UOMConvert(st.Temperature, enums.PropTemperature, s.currentUOM, s.desiredUOM)
	s.state.Humidity = st.Humidity
	s.state.BatteryLevel = uint8(st.Battery)

	if !firstSeen {
		s.updatesChan <- &device.StateUpdateData{State: s.state}
	}

	return s.state
}
