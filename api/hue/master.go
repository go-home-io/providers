package main

import (
	"encoding/json"

	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/gobwas/glob"
)

// Init plugin on master node.
func (e *HueEmulator) initMaster(data *api.InitDataAPI) error {
	_, chUpdate := data.FanOut.SubscribeDeviceUpdates()
	e.communicator.Subscribe(e.chCommands)
	go e.masterCycle(chUpdate, e.chCommands)
	return nil
}

// Master internal cycle. Waits for incoming messages
// either from FanOut or from worker.
func (e *HueEmulator) masterCycle(devUpdates chan *common.MsgDeviceUpdate, devCommands chan []byte) {
	for {
		select {
		case update := <-devUpdates:
			go e.processIncomingDeviceUpdate(update)
		case cmd, ok := <-devCommands:
			if !ok {
				return
			}
			go e.processDeviceCommands(cmd)
		}
	}
}

// Processes messages received from worker.
// If message is marked as discovery -- it's first load on a worker.
// We need to send all known devices states.
func (e *HueEmulator) processDeviceCommands(msg []byte) {
	cmd := &DeviceCommandMessage{}
	err := json.Unmarshal(msg, cmd)
	if err != nil {
		e.logger.Error("Received corrupted message", err)
		return
	}

	if cmd.IsDiscovery {
		for _, v := range e.devices {
			e.communicator.Publish(v)
		}

		return
	}

	g, err := glob.Compile(cmd.DeviceID)
	if err != nil {
		e.logger.Error("Failed to compile device regexp", err)
		return
	}

	a := make(map[string]interface{})
	data, err := json.Marshal(cmd.Attributes)
	if err != nil {
		e.logger.Error("Failed to encode command message", err)
		return
	}
	err = json.Unmarshal(data, &a)
	if err != nil {
		e.logger.Error("Failed to decode command message", err)
	}

	e.communicator.InvokeDeviceCommand(g, cmd.Command, a)
}

// Processes FanOut device updates, converts them to worker msg
// and send through service bus.
func (e *HueEmulator) processIncomingDeviceUpdate(msg *common.MsgDeviceUpdate) {
	e.Lock()
	defer e.Unlock()

	if helpers.SliceContainsString(e.unsupportedDevices, msg.ID) {
		return
	}

	var out *DeviceUpdateMessage

	out, ok := e.devices[msg.ID]
	if !ok {
		if !e.isMatch(msg) {
			e.unsupportedDevices = append(e.unsupportedDevices, msg.ID)
			return
		}

		out = &DeviceUpdateMessage{
			DeviceType: msg.Type,
			Name:       e.getDeviceName(msg),
			DeviceID:   msg.ID,
			State:      make(map[enums.Property]interface{}),
		}

		e.devices[msg.ID] = out
	}

	wasUpdated := false

	for k, v := range msg.State {
		if !enums.SliceContainsProperty(supportedProperties, k) {
			continue
		}

		out.State[k] = v
		wasUpdated = true
	}

	if wasUpdated {
		e.communicator.Publish(out)
	}
}

// Validates whether device matches filter regexps and its type is supported.
func (e *HueEmulator) isMatch(msg *common.MsgDeviceUpdate) bool {
	if !enums.SliceContainsDeviceType(e.Settings.types, msg.Type) {
		return false
	}

	for _, v := range e.Settings.devRegexp {
		if v.Match(msg.ID) {
			return true
		}
	}

	return false
}

// Returns either overwritten name or uses deviceID
func (e *HueEmulator) getDeviceName(msg *common.MsgDeviceUpdate) string {
	name, ok := e.Settings.NamesOverrides[msg.ID]
	if ok {
		return name
	}

	return msg.Name
}
