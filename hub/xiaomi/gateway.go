package main

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/vkorn/go-miio"
)

// Defines Xiaomi hub's LED light.
type xiaomiGateway struct {
	xiaomiDevice
	gateway *miio.Gateway
	state   *device.LightState
}

// GetSpec returns device spec.
func (g *xiaomiGateway) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropOn, enums.PropColor, enums.PropBrightness},
		SupportedCommands: []enums.Command{enums.CmdOn, enums.CmdOff, enums.CmdToggle,
			enums.CmdSetColor, enums.CmdSetBrightness},
	}
}

// On turns the light on.
func (g *xiaomiGateway) On() error {
	return g.gateway.On()
}

// Off turns the light off.
func (g *xiaomiGateway) Off() error {
	return g.gateway.Off()
}

// Toggle toggles the light state.
func (g *xiaomiGateway) Toggle() error {
	if !g.state.On {
		return g.On()
	}

	return g.Off()
}

// SetColor sets the LED color.
func (g *xiaomiGateway) SetColor(c common.Color) error {
	return g.gateway.SetColor(c.Color())
}

// SetBrightness sets the LED brightness.
func (g *xiaomiGateway) SetBrightness(br device.GradualBrightness) error {
	return g.gateway.SetBrightness(br.Value)
}

// Load is not used since hub is responsible for device init.
func (g *xiaomiGateway) Load() (*device.LightState, error) {
	return g.state, nil
}

// Update is not used since device pushes updates.
func (g *xiaomiGateway) Update() (*device.LightState, error) {
	return nil, nil
}

// SetScene is not supported.
func (g *xiaomiGateway) SetScene(common.String) error {
	return nil
}

// SetTransitionTime is not supported.
func (g *xiaomiGateway) SetTransitionTime(common.Int) error {
	return nil
}

// InternalUpdate performs internal state update in response of device messages.
func (g *xiaomiGateway) InternalUpdate(state interface{}, firstSeen bool) interface{} {
	s := state.(*miio.GatewayState)
	g.state.On = s.On
	g.state.BrightnessPercent = s.Brightness
	g.state.Color = common.NewColor(s.RGB)

	if !firstSeen {
		g.updatesChan <- &device.StateUpdateData{State: g.state}
	}

	return g.state
}
