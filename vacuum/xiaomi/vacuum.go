package main

import (
	"errors"
	"sync"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/vkorn/go-miio"
)

// XiaomiVacuum implements vacuum plugin.
type XiaomiVacuum struct {
	sync.Mutex

	Settings *Settings

	logger common.ILoggerProvider
	uom    enums.UOM

	vacuum *miio.Vacuum
	state  *device.VacuumState

	updateChan       chan *miio.DeviceUpdateMessage
	deviceUpdateChan chan *device.StateUpdateData
	stopChan         chan bool
}

// Init does nothing.
func (v *XiaomiVacuum) Init(data *device.InitDataDevice) error {
	miio.LOGGER = &xiaomiLogger{
		logger: data.Logger,
	}
	v.logger = data.Logger
	v.deviceUpdateChan = data.DeviceStateUpdateChan
	v.uom = data.UOM
	v.stopChan = make(chan bool)
	return nil
}

// Unload stops the device.
func (v *XiaomiVacuum) Unload() {
	v.stopChan <- true
	v.vacuum.Stop()
}

// GetName returns vacuum IP.
func (v *XiaomiVacuum) GetName() string {
	return v.Settings.IP
}

// GetSpec returns device specification.
func (v *XiaomiVacuum) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropVacStatus, enums.PropBatteryLevel, enums.PropArea,
			enums.PropDuration, enums.PropFanSpeed},
		SupportedCommands: []enums.Command{enums.CmdOn, enums.CmdOff, enums.CmdPause,
			enums.CmdDock, enums.CmdFindMe, enums.CmdSetFanSpeed},
	}
}

// Load starts vacuum monitoring.
func (v *XiaomiVacuum) Load() (*device.VacuumState, error) {
	v.state = &device.VacuumState{
		VacStatus: enums.VacUnknown,
	}
	vac, err := miio.NewVacuum(v.Settings.IP, v.Settings.Key)
	if err != nil {
		return nil, err
	}

	v.vacuum = vac
	v.updateChan = v.vacuum.UpdateChan
	go v.updates()
	go v.vacuum.UpdateStatus()
	return v.state, nil
}

// On starts cleaning.
func (v *XiaomiVacuum) On() error {
	if !v.vacuum.StartCleaning() {
		err := errors.New("miio error")
		v.logger.Error("Failed to turn vacuum on", err)
		return err
	}

	return nil
}

// Off stops cleaning.
func (v *XiaomiVacuum) Off() error {
	if !v.vacuum.StopCleaning() {
		err := errors.New("miio error")
		v.logger.Error("Failed to turn vacuum off", err)
		return err
	}

	return nil
}

// Pause pauses cleaning.
func (v *XiaomiVacuum) Pause() error {
	if !v.vacuum.PauseCleaning() {
		err := errors.New("miio error")
		v.logger.Error("Failed to pause vacuum", err)
		return err
	}

	return nil
}

// Dock sends vacuum to the dock.
func (v *XiaomiVacuum) Dock() error {
	if !v.vacuum.StopCleaningAndDock() {
		err := errors.New("miio error")
		v.logger.Error("Failed to dock vacuum", err)
		return err
	}

	return nil
}

// FindMe sends find me signal.
func (v *XiaomiVacuum) FindMe() error {
	if !v.vacuum.FindMe() {
		err := errors.New("miio error")
		v.logger.Error("Failed to send find me command to vacuum", err)
		return err
	}

	return nil
}

// SetFanSpeed sets fan speed.
func (v *XiaomiVacuum) SetFanSpeed(speed common.Percent) error {
	if !v.vacuum.SetFanPower(speed.Value) {
		err := errors.New("miio error")
		v.logger.Error("Failed to set vacuum fan speed", err)
		return err
	}

	return nil
}

// Update returns current state.
func (v *XiaomiVacuum) Update() (*device.VacuumState, error) {
	return v.state, nil
}

// Internal updates.
func (v *XiaomiVacuum) updates() {
	tick := time.Tick(60 * time.Second)
	for {
		select {
		case <-v.updateChan:
			go v.processUpdates()
		case <-tick:
			go v.vacuum.UpdateStatus()
		case <-v.stopChan:
			return
		}
	}
}

// Processing updates from the vacuum.
func (v *XiaomiVacuum) processUpdates() {
	v.Lock()
	defer v.Unlock()

	v.state.BatteryLevel = uint8(v.vacuum.State.Battery)
	if v.state.BatteryLevel > 100 {
		v.state.BatteryLevel = 100
	}

	v.state.Duration = v.vacuum.State.CleanTime
	v.state.FanSpeed = uint8(v.vacuum.State.FanPower)
	if v.state.FanSpeed > 100 {
		v.state.FanSpeed = 100
	}

	v.state.Area = float64(v.vacuum.State.CleanArea) * 0.0001
	v.state.Area = helpers.UOMConvert(v.state.Area, enums.PropArea, enums.UOMMetric, v.uom)

	v.state.VacStatus = enums.VacUnknown

	if v.vacuum.State.IsCleaning {
		v.state.VacStatus = enums.VacCleaning
	}

	switch v.vacuum.State.State {
	case miio.VacStateInitiating, miio.VacStateSleeping, miio.VacStateWaiting,
		miio.VacStateShuttingDown, miio.VacStateUpdating, miio.VacStateDocking:
		v.state.VacStatus = enums.VacDocked
	case miio.VacStateCleaning, miio.VacStateReturning, miio.VacStateSpot, miio.VacStateZone:
		v.state.VacStatus = enums.VacCleaning
	case miio.VacStatePaused:
		v.state.VacStatus = enums.VacPaused
	case miio.VacStateCharging:
		v.state.VacStatus = enums.VacCharging
	case miio.VacStateFull:
		v.state.VacStatus = enums.VacFull
	}

	if v.vacuum.State.Error == miio.VacErrorFull {
		v.state.VacStatus = enums.VacFull
	}

	v.deviceUpdateChan <- &device.StateUpdateData{
		State: v.state,
	}
}
