// +build linux,arm

package main

import (
	"fmt"
	"strconv"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/maciej/bme280"
	"golang.org/x/exp/io/i2c"
)

const (
	// log token for the ic2 device number.
	logI2CIDToken = "i2c_number"
)

// BME280Sensor defines bme280 environmental sensor.
type BME280Sensor struct {
	Settings *Settings
	logger   common.ILoggerProvider
	state    *device.SensorState
	uom      enums.UOM

	device *i2c.Device
	driver *bme280.Driver
}

// Init performs sensor load.
func (s *BME280Sensor) Init(data *device.InitDataDevice) error {
	s.logger = data.Logger
	s.uom = data.UOM

	d, err := i2c.Open(&i2c.Devfs{
		Dev: fmt.Sprintf("/dev/i2c-%d", s.Settings.DeviceID),
	}, s.Settings.Address)

	if err != nil {
		s.logger.Error("Failed to init I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		return err
	}

	s.device = d
	s.driver = bme280.New(s.device)
	s.state = &device.SensorState{
		SensorType: enums.SenTemperature,
	}
	return nil
}

// Unload stops the sensor polling.
func (s *BME280Sensor) Unload() {
	err := s.device.Close()
	if err != nil {
		s.logger.Error("Error closing I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
	}
}

// GetName returns the name.
func (s *BME280Sensor) GetName() string {
	return "bme280"
}

// GetSpec returns the sensor spec.
func (s *BME280Sensor) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropTemperature, enums.PropHumidity, enums.PropPressure},
		UpdatePeriod:        s.Settings.updateInterval,
	}
}

// Load performs plugin initial load.
func (s *BME280Sensor) Load() (*device.SensorState, error) {
	return s.Update()
}

// Unload retrieves current data.
func (s *BME280Sensor) Update() (*device.SensorState, error) {
	data, err := s.driver.Read()
	if err != nil {
		s.logger.Error("Error reading I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		return nil, err
	}

	s.state.Humidity = data.Humidity
	s.state.Temperature = helpers.UOMConvert(data.Temperature, enums.PropTemperature, enums.UOMMetric, s.uom)
	s.state.Humidity = helpers.UOMConvert(data.Humidity, enums.PropHumidity, enums.UOMMetric, s.uom)
	return s.state, nil
}
