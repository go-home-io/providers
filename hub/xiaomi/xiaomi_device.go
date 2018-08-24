//go:generate enumer -type=xiaomiDeviceType -transform=snake

package main

import (
	"fmt"

	"github.com/go-home-io/server/plugins/device"
)

// Defines device type.
type xiaomiDeviceType int

const (
	// Hub device.
	gateway xiaomiDeviceType = iota
	// Switch device.
	button
	// Magnet device.
	magnet
	// Temperature-humidity sensor.
	temperature
	// Motion sensor.
	motion
)

// IInternalDevice defines internal Xiaomi device.
type IInternalDevice interface {
	GetName() string
	Init(data *device.InitDataDevice) error
	Unload()
	GetID() string
}

// IXiaomiDevice defines generic Xiaomi device.
type IXiaomiDevice interface {
	IInternalDevice
	InternalUpdate(state interface{}, firstSeen bool) interface{}
}

// Defines Xiaomi device.
type xiaomiDevice struct {
	deviceID    string
	deviceType  xiaomiDeviceType
	updatesChan chan *device.StateUpdateData
}

// Init performs initial device init.
func (d *xiaomiDevice) Init(data *device.InitDataDevice) error {
	d.updatesChan = data.DeviceStateUpdateChan
	return nil
}

// Stops the device. Since only hub needs to be stopped, this method is empty.
func (d *xiaomiDevice) Unload() {
}

// GetName returns the device name.
func (d *xiaomiDevice) GetName() string {
	return fmt.Sprintf("%s.%s", d.deviceType.String(), d.deviceID)
}

// GetID returns the device ID.
func (d *xiaomiDevice) GetID() string {
	return d.deviceID
}
