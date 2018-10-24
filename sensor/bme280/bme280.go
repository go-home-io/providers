package main

import (
	"strconv"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/bmxx80"
	"periph.io/x/periph/host"
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
	update   chan *device.StateUpdateData

	bus     i2c.BusCloser
	device  *bmxx80.Dev
	data    <-chan physic.Env
	closeCh chan bool
}

// Init performs sensor load.
func (s *BME280Sensor) Init(data *device.InitDataDevice) error {
	s.logger = data.Logger
	s.uom = data.UOM
	s.update = data.DeviceStateUpdateChan

	_, err := host.Init()
	if err != nil {
		s.logger.Error("Failed to init periph host", err)
		return errors.Wrap(err, "host init failed")
	}

	bus, err := i2creg.Open(strconv.Itoa(s.Settings.DeviceID))
	s.bus = bus

	if err != nil {
		s.logger.Error("Failed to init I2C bus", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		errBus := s.bus.Close()
		if errBus != nil {
			s.logger.Error("Error closing I2C bus", errBus, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		}

		return errors.Wrap(err, "bus init failed")
	}

	dev, err := bmxx80.NewI2C(s.bus, uint16(s.Settings.Address), &bmxx80.DefaultOpts)
	if err != nil {
		s.logger.Error("Failed to init I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		return errors.Wrap(err, "device init failed")
	}

	s.device = dev
	s.state = &device.SensorState{
		SensorType: enums.SenTemperature,
	}

	return nil
}

// Unload stops the sensor polling.
func (s *BME280Sensor) Unload() {
	s.closeCh <- true
}

// GetName returns the name.
func (s *BME280Sensor) GetName() string {
	return "bme280"
}

// GetSpec returns the sensor spec.
func (s *BME280Sensor) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropTemperature, enums.PropHumidity, enums.PropPressure},
	}
}

// Load performs plugin initial load.
func (s *BME280Sensor) Load() (*device.SensorState, error) {
	s.closeCh = make(chan bool, 1)
	ch, err := s.device.SenseContinuous(s.Settings.updateInterval)
	if err != nil {
		s.logger.Error("Error reading I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
		return nil, errors.Wrap(err, "i2c read failed")
	}

	s.data = ch
	return s.state, nil
}

// Update is not used since plugin is using a callback.
func (s *BME280Sensor) Update() (*device.SensorState, error) {
	return s.state, nil
}

// Waits for update messages.
func (s *BME280Sensor) updates() {
	for {
		select {
		case data := <-s.data:
			s.doUpdate(data)
		case <-s.closeCh:
			s.close()
			return
		}
	}
}

// Performs state update.
func (s *BME280Sensor) doUpdate(data physic.Env) {
	s.state.Humidity = float64(data.Humidity)
	s.state.Temperature = helpers.UOMConvert(float64(data.Temperature), enums.PropTemperature, enums.UOMMetric, s.uom)
	// Pressure is reported in kPa which is 10 mbar.
	s.state.Pressure = helpers.UOMConvert(float64(data.Pressure)*10.0, enums.PropPressure, enums.UOMMetric, s.uom)

	s.update <- &device.StateUpdateData{
		State: s.state,
	}
}

// Closes device.
func (s *BME280Sensor) close() {
	err := s.device.Halt()
	if err != nil {
		s.logger.Error("Error closing I2C device", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
	}
	err = s.bus.Close()
	if err != nil {
		s.logger.Error("Error closing I2C bus", err, logI2CIDToken, strconv.Itoa(s.Settings.DeviceID))
	}
}
