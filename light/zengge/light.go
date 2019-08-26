package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/vikstrous/zengge-lightcontrol/control"
	"github.com/vikstrous/zengge-lightcontrol/local"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
)

// ZenggeLight describes Zengge device.
type ZenggeLight struct {
	Settings *Settings
	State    *device.LightState

	logger common.ILoggerProvider
	spec   *device.Spec
}

// Init performs initial plugin init.
func (z *ZenggeLight) Init(data *device.InitDataDevice) error {
	z.logger = data.Logger
	z.State = &device.LightState{
		Color: common.Color{},
	}
	z.spec = &device.Spec{
		UpdatePeriod:           z.Settings.pollingInterval,
		SupportedCommands:      []enums.Command{enums.CmdOn, enums.CmdOff, enums.CmdToggle, enums.CmdSetColor},
		SupportedProperties:    []enums.Property{enums.PropOn, enums.PropColor},
		PostCommandDeferUpdate: 500 * time.Millisecond,
	}

	return nil
}

// Unload is not used, since plugin doesn't keep any open connections.
func (z *ZenggeLight) Unload() {
}

// GetName returns device name.
// IP is used by default.
func (z *ZenggeLight) GetName() string {
	return z.Settings.LightIP
}

// GetSpec returns device specs.
func (z *ZenggeLight) GetSpec() *device.Spec {
	return z.spec
}

// Input is not used.
func (z *ZenggeLight) Input(common.Input) error {
	return nil
}

// Load performs initial device connection and state pulls.
func (z *ZenggeLight) Load() (*device.LightState, error) {
	transport, err := local.NewTransport(fmt.Sprintf("%s:%d", z.Settings.LightIP, devicePort))
	if err != nil {
		return nil, errors.Wrap(err, "tcp transport failed")
	}
	defer transport.Close()
	controller := &control.Controller{Transport: transport}
	internalState, err := controller.GetState()

	if err != nil {
		return nil, errors.Wrap(err, "get state failed")
	}

	z.logger.Info("Successfully connected to Zengge device", common.LogDeviceHostToken, z.Settings.LightIP)

	z.updateState(internalState)

	return z.State, nil
}

// On makes an attempt to turn device on.
func (z *ZenggeLight) On() error {
	return z.setPower(true)
}

// Off makes an attempt to turn device off.
func (z *ZenggeLight) Off() error {
	return z.setPower(false)
}

// Toggle makes an attempt to toggle device state.
func (z *ZenggeLight) Toggle() error {
	if z.State.On {
		return z.Off()
	}

	return z.On()
}

// Update performs call for the update.
func (z *ZenggeLight) Update() (*device.LightState, error) {
	transport, err := local.NewTransport(fmt.Sprintf("%s:%d", z.Settings.LightIP, devicePort))
	if err != nil {
		return nil, errors.Wrap(err, "tcp transport failed")
	}
	defer transport.Close()
	controller := &control.Controller{Transport: transport}
	internalState, err := controller.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "get state failed")
	}

	z.updateState(internalState)
	return z.State, nil
}

// SetBrightness is not supported.
func (z *ZenggeLight) SetBrightness(device.GradualBrightness) error {
	return nil
}

// SetScene is not supported.
func (z *ZenggeLight) SetScene(string common.String) error {
	return nil
}

// SetColor makes an attempt to change device color.
func (z *ZenggeLight) SetColor(color common.Color) error {
	if !z.State.On {
		err := z.On()
		if err != nil {
			return errors.Wrap(err, "set color failed")
		}
	}

	c := control.Color{
		R:    color.R,
		G:    color.G,
		B:    color.B,
		UseW: false,
	}

	transport, err := local.NewTransport(fmt.Sprintf("%s:%d", z.Settings.LightIP, devicePort))
	if err != nil {
		return errors.Wrap(err, "tcp transport failed")
	}
	defer transport.Close()
	controller := &control.Controller{Transport: transport}

	err = controller.SetColor(c)

	if err != nil {
		z.logger.Error("Failed to set color of Zengge device", err, common.LogDeviceHostToken, z.Settings.LightIP)
		return errors.Wrap(err, "set color failed")
	}

	return nil
}

// SetTransitionTime is not supported.
func (z *ZenggeLight) SetTransitionTime(common.Int) error {
	return nil
}

// Updates internal state.
func (z *ZenggeLight) updateState(internalState *control.State) {
	z.State.On = internalState.IsOn
	z.State.Color.R = internalState.Color.R
	z.State.Color.G = internalState.Color.G
	z.State.Color.B = internalState.Color.B
}

// Makes an attempt to change power state.
func (z *ZenggeLight) setPower(isOn bool) error {
	transport, err := local.NewTransport(fmt.Sprintf("%s:%d", z.Settings.LightIP, devicePort))
	if err != nil {
		return errors.Wrap(err, "tcp transport init failed")
	}
	defer transport.Close()
	controller := &control.Controller{Transport: transport}

	err = controller.SetPower(isOn)
	if err != nil {
		z.logger.Error("Failed to set power of Zengge device", err,
			common.LogDeviceHostToken, z.Settings.LightIP)
		return errors.Wrap(err, "set power failed")
	}

	return nil
}
