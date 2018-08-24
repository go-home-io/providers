package main

import (
	"time"

	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/vkorn/go-miio"
)

// Defines Xiaomi switch.
type xiaomiButton struct {
	xiaomiDevice
	state *device.SensorState
}

// GetSpec returns device spec.
func (b *xiaomiButton) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropClick, enums.PropDoubleClick,
			enums.PropPress, enums.PropBatteryLevel, enums.PropSensorType},
		SupportedCommands: []enums.Command{},
	}
}

// Load is not used since hub is responsible for device init.
func (b *xiaomiButton) Load() (*device.SensorState, error) {
	return b.state, nil
}

// Update is not used since device pushes updates.
func (b *xiaomiButton) Update() (*device.SensorState, error) {
	return nil, nil
}

// InternalUpdate performs internal state update in response of device messages.
func (b *xiaomiButton) InternalUpdate(state interface{}, firstSeen bool) interface{} {
	s := state.(*miio.SwitchState)
	b.state.Click = false
	b.state.DoubleClick = false
	b.state.Press = false
	b.state.BatteryLevel = uint8(s.Battery)

	switch s.Click {
	case miio.ClickSingle:
		b.state.Click = true
		go b.resetState()
	case miio.ClickDouble:
		b.state.DoubleClick = true
		go b.resetState()
	case miio.ClickLongRelease:
		b.state.Press = true
		go b.resetState()
	}

	if !firstSeen {
		b.updatesChan <- &device.StateUpdateData{State: b.state}
	}

	return b.state
}

// Resets click state after five seconds since switches are not reporting this.
func (b *xiaomiButton) resetState() {
	time.Sleep(5 * time.Second)
	b.state.Click = false
	b.state.DoubleClick = false
	b.state.Press = false
	b.updatesChan <- &device.StateUpdateData{State: b.state}
}
