package main

import (
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"periph.io/x/periph/host/sysfs"
)

const (
	// Bus ID token.
	logBusIDToken = "bus_number"
	// Device ID token.
	logDeviceIDToken = "spi_number"
)

// RC522Sensor defines rc522 RFID sensor.
type RC522Sensor struct {
	Settings *Settings
	logger   common.ILoggerProvider
	state    *device.SensorState

	update chan *device.StateUpdateData
	bus    *sysfs.SPI
	device *mfrc522.Dev
	stop   bool
}

// Init performs sensor load.
func (s *RC522Sensor) Init(data *device.InitDataDevice) error {
	s.update = data.DeviceStateUpdateChan
	s.logger = data.Logger

	b, err := sysfs.NewSPI(s.Settings.BusID, s.Settings.DeviceID)
	if err != nil {
		return errors.Wrap(err, "bus init failed")
	}

	s.bus = b
	// TODO: Need a PR
	//s.device.SetAntennaGain(s.Settings.AntennaGain)

	d, err := mfrc522.NewSPI(s.bus, s.Settings.rstGPIO, s.Settings.irqGPIO)
	if err != nil {
		s.closeBus()
		return errors.Wrap(err, "device init failed")
	}

	s.device = d
	s.state = &device.SensorState{
		On:         false,
		SensorType: enums.SenPresence,
	}
	return nil
}

// Unload stops the sensor polling.
func (s *RC522Sensor) Unload() {
	// TODO: Need a PR to stop Wait cycle
	err := s.device.Halt()
	if err != nil {
		s.logger.Error("Failed to close SPI device", err,
			logBusIDToken, strconv.Itoa(s.Settings.BusID), logDeviceIDToken, strconv.Itoa(s.Settings.DeviceID))
	}

	s.closeBus()
	s.stop = true
}

// GetName returns the name.
func (s *RC522Sensor) GetName() string {
	return "rc522"
}

// GetSpec returns the sensor spec.
func (s *RC522Sensor) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropOn, enums.PropSensorType, enums.PropUser},
	}
}

// Input is not used.
func (s *RC522Sensor) Input(common.Input) error {
	return nil
}

// Load performs plugin initial load.
func (s *RC522Sensor) Load() (*device.SensorState, error) {
	go s.readData()
	s.state.On = false
	return s.state, nil
}

// Update is not used since plugin is using a callback.
func (s *RC522Sensor) Update() (*device.SensorState, error) {
	return s.state, nil
}

// Data poll from the device.
func (s *RC522Sensor) readData() {
	for {
		// We could've received stop command while sleeping.
		// No need for additional panic.
		if s.stop {
			return
		}

		data := s.waitRead()
		if s.stop {
			return
		}

		if 0 == len(data) || nil == data {
			continue
		}

		s.dataReceived(data)
		// We don't want to record every single tick
		time.Sleep(1 * time.Second)
	}
}

// Closes SPI bus
func (s *RC522Sensor) closeBus() {
	err := s.bus.Close()
	if err != nil {
		s.logger.Error("Failed to close SPI bus", err,
			logBusIDToken, strconv.Itoa(s.Settings.BusID), logDeviceIDToken, strconv.Itoa(s.Settings.DeviceID))
	}
}

// Analyzes received data.
func (s *RC522Sensor) dataReceived(data []byte) {
	for k, v := range s.Settings.users {
		if reflect.DeepEqual(v, data) {
			s.logger.Info("Detected valid RFID",
				common.LogUserNameToken, k, logBusIDToken, strconv.Itoa(s.Settings.BusID),
				logDeviceIDToken, strconv.Itoa(s.Settings.DeviceID))
			s.state.User = k
			s.state.On = true

			s.update <- &device.StateUpdateData{
				State: s.state,
			}

			go s.resetState()
			return
		}
	}

	s.logger.Warn("Detected RFID but no valida data is present",
		logBusIDToken, strconv.Itoa(s.Settings.BusID), logDeviceIDToken, strconv.Itoa(s.Settings.DeviceID))
}

// Resets sensor state after 10 seconds.
func (s *RC522Sensor) resetState() {
	time.Sleep(10 * time.Second)
	s.state.On = false
	s.update <- &device.StateUpdateData{
		State: s.state,
	}
}

// Waits for the data to become available.
func (s *RC522Sensor) waitRead() []byte {
	for {
		data, err := s.device.ReadCard(byte(commands.PICC_AUTHENT1B),
			s.Settings.Sector, s.Settings.Block, s.Settings.encKey)
		if s.stop {
			return nil
		}

		if err != nil {
			continue
		}

		return data
	}
}
