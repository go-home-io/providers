package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/amimof/huego"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
)

// HueLight describes light or group resource, exposed by HUE bridge.
type HueLight struct {
	ID         string
	Bridge     *huego.Bridge
	InternalID int
	IsNew      bool

	IsGroup bool
	Light   huego.Light
	Group   huego.Group

	state  *device.LightState
	spec   *device.Spec
	logger common.ILoggerProvider

	patchedScenes map[string]string

	sharedObjects *sharedObjects
	updateChan    chan *device.StateUpdateData
}

// Init saves data only since hub is responsible for all updates.
func (h *HueLight) Init(data *device.InitDataDevice) error {
	h.logger = data.Logger
	h.updateChan = data.DeviceStateUpdateChan
	return nil
}

// Load is not used since hub is responsible for device init.
func (h *HueLight) Load() (*device.LightState, error) {
	return nil, nil
}

// Unload is not used since hub is responsible for device updates.
func (h *HueLight) Unload() {
}

// GetName returns device name.
// If it's a group, name will be prefixed by "group".
func (h *HueLight) GetName() string {
	if !h.IsGroup {
		return h.ID
	}

	return fmt.Sprintf("group.%s", h.ID)
}

// GetSpec returns device spec.
func (h *HueLight) GetSpec() *device.Spec {
	return h.spec
}

// SetBrightness makes an attempt to change device brightness.
func (h *HueLight) SetBrightness(percent device.GradualBrightness) error {
	if percent.TransitionSeconds > 0 {
		h.changeBrightnessOverTime(percent)
		return nil
	}

	val := uint8(float32(percent.Value) * float32(brightnessMax) / 100.0)

	var err error

	if h.IsGroup {
		err = h.Group.Bri(val)
	} else {
		err = h.Light.Bri(val)
	}

	if err != nil {
		h.logger.Error("Failed to set HUE brightness", err, common.LogDeviceNameToken, h.ID)
	} else {
		h.performActualUpdate(true)
	}

	return err
}

// On makes an attempt to turn device on.
// nolint: dupl
func (h *HueLight) On() error {
	var err error
	if h.IsGroup {
		err = h.Group.On()
	} else {
		err = h.Light.On()
	}

	if err != nil {
		h.logger.Error("Failed to turn on HUE", err, common.LogDeviceNameToken, h.ID)
	} else {
		h.state.On = !h.state.On
		h.performActualUpdate(true)
	}

	return err
}

// Off makes an attempt to turn device off.
// nolint: dupl
func (h *HueLight) Off() error {
	var err error
	if h.IsGroup {
		err = h.Group.Off()
	} else {
		err = h.Light.Off()
	}

	if err != nil {
		h.logger.Error("Failed to turn off HUE", err, common.LogDeviceNameToken, h.ID)
	} else {
		h.state.On = !h.state.On
		h.performActualUpdate(true)
	}

	return err
}

// Toggle makes an attempt to toggle device state.
func (h *HueLight) Toggle() error {
	var err error
	if h.IsGroup {
		if h.state.On {
			err = h.Group.Off()
		} else {
			err = h.Group.On()
		}
	} else {
		if h.state.On {
			err = h.Light.Off()
		} else {
			err = h.Light.On()
		}
	}

	if err != nil {
		h.logger.Error("Failed to toggle HUE", err, common.LogDeviceNameToken, h.ID)
	} else {
		h.state.On = !h.state.On
		h.performActualUpdate(true)
	}

	return err
}

// SetScene makes an attempt to set device scene.
// Scenes are supported by groups only.
func (h *HueLight) SetScene(str common.String) error {
	if !h.IsGroup {
		return nil
	}

	if id, ok := h.patchedScenes[str.Value]; ok {
		if scene, ok := h.sharedObjects.scenes[id]; ok {
			err := scene.Scene.Recall(h.InternalID)
			if err != nil {
				return err
			}

			h.performActualUpdate(true)
		}
	}
	h.logger.Warn("Failed to find HUE scene", common.LogDeviceNameToken, h.ID, "scene", str.Value)
	return errors.New("scene not found")
}

// Update returns current device state.
func (h *HueLight) Update() (*device.LightState, error) {
	return h.state, nil
}

// SetTransitionTime makes an attempt to set transition time.
func (h *HueLight) SetTransitionTime(int common.Int) error {
	var err error
	if h.IsGroup {
		err = h.Group.TransitionTime(uint16(int.Value))
	} else {
		err = h.Light.TransitionTime(uint16(int.Value))
	}

	return err
}

// SetColor makes an attempt to change device color.
func (h *HueLight) SetColor(color common.Color) error {
	var err error
	x, y := rgb2cie(color)
	if h.IsGroup {
		err = h.Group.Xy([]float32{x, y})

	} else {
		err = h.Light.Xy([]float32{x, y})
	}

	if err != nil {
		h.logger.Error("Failed to set HUE color", err, common.LogDeviceNameToken, h.ID)
		return err
	}

	h.performActualUpdate(true)
	return nil
}

// Gradually changes device brightness over time.
func (h *HueLight) changeBrightnessOverTime(percent device.GradualBrightness) {
	desiredValue := uint8(float32(percent.Value) * float32(brightnessMax) / 100.0)

	var currentValue uint8
	if h.IsGroup {
		currentValue = uint8(h.Group.State.Bri)
	} else {
		currentValue = uint8(h.Light.State.Bri)
	}

	step := (float32(desiredValue) - float32(currentValue)) /
		float32(percent.TransitionSeconds*overtimeBrightnessStepsPerSecond)
	max := int(percent.TransitionSeconds)
	curVal := uint8(0)

	go func() {
		for ii := 0; ii < max*overtimeBrightnessStepsPerSecond; ii++ {
			nextVal := uint8(float32(currentValue) + step*float32(ii+1))
			if nextVal != curVal {
				if h.IsGroup {
					h.Group.Bri(nextVal)
				} else {
					h.Light.Bri(nextVal)
				}
			}

			curVal = nextVal
			time.Sleep(1000 * time.Millisecond / overtimeBrightnessStepsPerSecond)
		}
		if h.IsGroup {
			h.Group.Bri(desiredValue)
		} else {
			h.Light.Bri(desiredValue)
		}
		h.performActualUpdate(false)

	}()
}

// Updates device state.
func (h *HueLight) setState(state *huego.State) {
	h.spec = &device.Spec{
		UpdatePeriod:        h.sharedObjects.settings.pollingInterval,
		SupportedCommands:   []enums.Command{enums.CmdOn, enums.CmdOff, enums.CmdToggle, enums.CmdSetBrightness},
		SupportedProperties: []enums.Property{enums.PropOn, enums.PropBrightness},
	}

	if state.Hue > 0 {
		h.spec.SupportedCommands = append(h.spec.SupportedCommands, enums.CmdSetColor)
		h.spec.SupportedProperties = append(h.spec.SupportedProperties, enums.PropColor)
	}

	if h.IsGroup {
		h.spec.SupportedCommands = append(h.spec.SupportedCommands, enums.CmdSetScene)
		h.spec.SupportedProperties = append(h.spec.SupportedProperties, enums.PropScenes)
	}

	h.processUpdate(state)

	if h.state.TransitionTime > 0 {
		h.spec.SupportedCommands = append(h.spec.SupportedCommands, enums.CmdSetTransitionTime)
		h.spec.SupportedProperties = append(h.spec.SupportedProperties, enums.PropTransitionTime)
	}
}

// Processes received HUE state.
func (h *HueLight) processUpdate(state *huego.State) {
	h.state.On = state.On
	h.state.TransitionTime = state.TransitionTime
	h.state.BrightnessPercent = uint8((float32(state.Bri) * 100.0) / float32(brightnessMax))

	if len(state.Xy) > 1 {
		h.state.Color = cie2rgb(state.Xy[0], state.Xy[1], float32(h.state.BrightnessPercent))
	}

	if h.IsGroup {
		h.state.Scenes = h.pickScenes()
	}
}

// Iterates through all known to bridge states and
// select which groups have same lights.
// Scenes are supported by groups only.
func (h *HueLight) pickScenes() []string {
	if !h.IsGroup {
		return nil
	}

	h.patchedScenes = make(map[string]string)
	h.sharedObjects.Lock()
	defer h.sharedObjects.Unlock()

	scenes := make([]string, 0)
	for _, v := range h.sharedObjects.scenes {
		if helpers.SliceEqualsString(v.Scene.Lights, h.Group.Lights) {

			originalName := v.Scene.Name
			finalName := originalName
			for ii := 1; ii <= 10; ii++ {
				if _, ok := h.patchedScenes[finalName]; ok {
					finalName = fmt.Sprintf("%s (%d)", originalName, ii)
				} else {
					h.patchedScenes[finalName] = v.Scene.ID
					scenes = append(scenes, finalName)
					break
				}
			}
		}
	}
	return scenes
}

// Performs call to device API to forcefully pull an update.
// Method is used when any state-updates command was invoked
// to sync with internal state data.
func (h *HueLight) performActualUpdate(skipPublishing bool) {
	if h.IsGroup {
		g, _ := h.Bridge.GetGroup(h.InternalID)
		h.setState(g.State)

		for _, i := range h.Group.Lights {
			id, err := strconv.Atoi(i)
			if err == nil {
				h.sharedObjects.internalLightUpdates <- id
			}
		}
	} else {
		l, _ := h.Bridge.GetLight(h.InternalID)
		h.setState(l.State)
	}

	if skipPublishing {
		return
	}

	h.updateChan <- &device.StateUpdateData{
		State: h.state,
	}
}
