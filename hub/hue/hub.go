package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/amimof/huego"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/pkg/errors"
)

// HueHub describes HUE hub state.
type HueHub struct {
	logger common.ILoggerProvider
	secret common.ISecretProvider
	bridge *huego.Bridge

	lights map[int]*HueLight
	groups map[int]*HueLight

	state *device.HubState
	spec  *device.Spec

	sharedObjects *sharedObjects
}

// Describes known to hub scene.
type hueScene struct {
	ID    string
	Scene huego.Scene
}

// Helper object for sharing data between hub and devices.
type sharedObjects struct {
	sync.Mutex
	scenes   map[string]hueScene
	settings *Settings

	internalLightUpdates chan int
}

// Init performs initial plugin init.
func (h *HueHub) Init(data *device.InitDataDevice) error {
	h.logger = data.Logger
	h.secret = data.Secret
	h.lights = make(map[int]*HueLight)
	h.groups = make(map[int]*HueLight)
	h.sharedObjects.scenes = make(map[string]hueScene)
	h.state = &device.HubState{}
	h.spec = &device.Spec{
		UpdatePeriod:        h.sharedObjects.settings.pollingInterval,
		SupportedCommands:   []enums.Command{},
		SupportedProperties: []enums.Property{enums.PropNumDevices},
	}
	return nil
}

// Load makes an attempt to authenticate over HUE hub.
// If IP and token are not provided, discovery attempt is made.
// By default, to register a new user, "go-home" is used.
// Plugin supports go-home Secret store and in case of
// success registration, stores received token into it.
func (h *HueHub) Load() (*device.HubLoadResult, error) {
	if err := h.authenticate(); err != nil {
		return nil, err
	}

	h.updateState()

	result := &device.HubLoadResult{
		State:   h.state,
		Devices: make([]*device.DiscoveredDevices, 0),
	}

	result.Devices = append(result.Devices, getNew(h.lights)...)
	result.Devices = append(result.Devices, getNew(h.groups)...)
	h.markAsOld()

	go h.internalUpdates()

	return result, nil
}

// Update makes an attempt to pull HUE hub for the new states.
func (h *HueHub) Update() (*device.HubLoadResult, error) {
	newDevices := getNew(h.lights)
	newDevices = append(newDevices, getNew(h.groups)...)
	h.markAsOld()

	return &device.HubLoadResult{
		State:   h.state,
		Devices: newDevices,
	}, nil
}

// Unload handles plugin unload.
func (h *HueHub) Unload() {
	close(h.sharedObjects.internalLightUpdates)
}

// GetName returns deice name.
// Hub IP address is used.
func (h *HueHub) GetName() string {
	return strings.NewReplacer("https://", "", "http://", "").Replace(h.bridge.Host)
}

// GetSpec returns hub specs.
func (h *HueHub) GetSpec() *device.Spec {
	return h.spec
}

// Pulls hub for a new devices.
func getNew(devices map[int]*HueLight) []*device.DiscoveredDevices {
	newDevices := make([]*device.DiscoveredDevices, 0)
	for _, v := range devices {
		if !v.IsNew {
			continue
		}

		newDevices = append(newDevices, &device.DiscoveredDevices{
			Interface: v,
			State:     v.state,
			Type:      enums.DevLight,
		})
	}

	return newDevices
}

// Internal updates cycle.
func (h *HueHub) internalUpdates() {
	for id := range h.sharedObjects.internalLightUpdates {
		for _, v := range h.lights {
			if v.InternalID == id {
				v.performActualUpdate(false)
				break // nolint: megacheck_provider
			}
		}
	}
}

// Performs authentication request over hub.
func (h *HueHub) authenticate() error {
	var err error
	if h.sharedObjects.settings.BridgeIP == "" {
		err = h.registerOrLogin()
		if err != nil {
			return errors.Wrap(err, "auth failed")
		}
	} else {
		h.bridge = huego.New(h.sharedObjects.settings.BridgeIP,
			h.sharedObjects.settings.Token).Login(h.sharedObjects.settings.Token)
	}

	_, err = h.bridge.GetConfig()
	if err != nil {
		return errors.New("failed to communicate to HUE hub")
	}

	h.logger.Info("Successfully authenticated against HUE hub", common.LogDeviceHostToken, h.bridge.Host)
	return nil
}

// Performs an attempt to login over a hub or
// to register a new user.
func (h *HueHub) registerOrLogin() error {
	var err error
	h.bridge, err = huego.Discover()
	if err != nil {
		h.logger.Error("HUE discovery failed", err)
		return errors.Wrap(err, "discovery failed")
	}

	secretName := fmt.Sprintf("hue-hub-%s", h.bridge.Host)
	needToSaveSecret := false

	if "" == h.sharedObjects.settings.Token {
		h.sharedObjects.settings.Token, err = h.secret.Get(secretName)
		if err != nil {
			h.logger.Info("No secrets present for hub, using default user for registration",
				common.LogDeviceHostToken, h.bridge.Host)
			h.sharedObjects.settings.Token = "go-home"
			needToSaveSecret = true
		} else {
			h.bridge = h.bridge.Login(h.sharedObjects.settings.Token)
			return nil
		}
	}

	h.logger.Info("Successfully discovered HUE bridge", common.LogDeviceHostToken, h.bridge.Host)
	user, err := h.bridge.CreateUser(h.sharedObjects.settings.Token)
	if err != nil {
		h.logger.Error("HUE registration failed, make sure link button was pressed", err,
			common.LogDeviceHostToken, h.bridge.Host)
		return errors.Wrap(err, "registration failed")
	}

	h.logger.Info("Successfully registered a new USER", common.LogDeviceHostToken, h.bridge.Host)
	if needToSaveSecret {
		if err = h.secret.Set(secretName, user); err != nil {
			h.logger.Error("Failed to save secret, make sure to save token manually", err,
				common.LogDeviceHostToken, h.bridge.Host, "token", user)
		}
	}

	h.bridge = h.bridge.Login(user)

	return nil
}

// Updates devices states.
func (h *HueHub) updateState() {
	h.loadScenes()

	for _, v := range h.sharedObjects.settings.load {
		switch v {
		case ResourceLights:
			h.loadLights()
		case ResourceGroups:
			h.loadGroups()
		}
	}

	h.state.NumDevices = len(h.lights) + len(h.groups)
}

// Helper method to mark loaded device as known.
func (h *HueHub) markAsOld() {
	for _, v := range h.lights {
		v.IsNew = false
	}

	for _, v := range h.groups {
		v.IsNew = false
	}
}

// Performs call to HUE hub to query known scenes.
func (h *HueHub) loadScenes() {
	h.sharedObjects.Lock()
	defer h.sharedObjects.Unlock()

	h.sharedObjects.scenes = make(map[string]hueScene)
	scenes, err := h.bridge.GetScenes()
	if err != nil {
		return
	}

	for _, v := range scenes {
		scene := hueScene{
			ID:    v.ID,
			Scene: v,
		}

		h.sharedObjects.scenes[scene.ID] = scene
	}
}

// Performs call to HUE hub to query known groups.
func (h *HueHub) loadGroups() {
	groups, err := h.bridge.GetGroups()
	if err != nil {
		h.logger.Error("Failed to load HUE groups", err)
		return
	}

	for _, g := range groups {
		if _, ok := h.groups[g.ID]; ok {
			continue
		}

		newG := HueLight{
			Bridge:        h.bridge,
			InternalID:    g.ID,
			ID:            g.Name,
			state:         &device.LightState{},
			IsNew:         true,
			IsGroup:       true,
			Group:         g,
			sharedObjects: h.sharedObjects,
		}
		newG.setState(g.State)

		h.groups[g.ID] = &newG
	}
}

// Performs call to HUE hub to query known lights.
func (h *HueHub) loadLights() {
	lights, err := h.bridge.GetLights()
	if err != nil {
		h.logger.Error("Failed to load HUE lights", err)
		return
	}

	for _, l := range lights {
		if !l.State.Reachable {
			h.logger.Debug("One of the HUE lights is unreachable", common.LogDeviceNameToken, l.Name)
			continue
		}

		if _, ok := h.lights[l.ID]; ok {
			continue
		}

		newL := HueLight{
			Bridge:        h.bridge,
			InternalID:    l.ID,
			ID:            l.Name,
			state:         &device.LightState{},
			IsNew:         true,
			IsGroup:       false,
			Light:         l,
			sharedObjects: h.sharedObjects,
		}
		newL.setState(l.State)

		h.lights[l.ID] = &newL
	}
}
