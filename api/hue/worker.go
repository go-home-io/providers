package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/julienschmidt/httprouter"
)

// Init plugin on worker node.
func (e *HueEmulator) initWorker(*api.InitDataAPI) error {
	e.upnp = &discoverUPNP{
		logger:     e.logger,
		advAddress: e.Settings.AdvAddress,
	}

	// It was validated before plugin's load
	parts := strings.Split(e.Settings.AdvAddress, ":")
	port, _ := strconv.Atoi(parts[1]) // nolint: gosec

	l, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	})

	if err != nil {
		return err
	}

	e.listener = l
	err = e.upnp.Start()
	if err != nil {
		return err
	}

	router := httprouter.New()
	router.GET("/upnp/setup.xml", e.upnp.Setup)
	router.GET("/api/:userID", e.getDevices)
	router.PUT("/api/:userID/lights/:lightID/state", e.setDeviceState)
	router.GET("/api/:userID/lights/:lightID", e.getDeviceInfo)

	go func() {
		err := http.Serve(e.listener, router)
		if err != nil {
			return
		}
	}()
	err = e.communicator.Subscribe(e.chCommands)
	if err != nil {
		return err
	}

	go e.workerCycle(e.chCommands)
	e.communicator.Publish(&DeviceCommandMessage{
		IsDiscovery: true,
	})
	return nil
}

// Worker internal bus cycle. Waits for incoming devices updates.
func (e *HueEmulator) workerCycle(devUpdates chan []byte) {
	for msg := range devUpdates {
		go e.processDeviceUpdate(msg)
	}
}

// Processes incoming device update message.
func (e *HueEmulator) processDeviceUpdate(msg []byte) {
	e.Lock()
	defer e.Unlock()

	update := &DeviceUpdateMessage{}
	err := json.Unmarshal(msg, update)

	if err != nil {
		e.logger.Error("Received corrupted message from master", err)
		return
	}

	var dHash string
	old, ok := e.devices[update.DeviceID]
	if ok {
		dHash = old.internalHash
	} else {
		dHash = hash(update.DeviceID)
	}

	e.devices[update.DeviceID] = update
	e.devices[update.DeviceID].internalHash = dHash
}

// Responds to GetAllLights API.
func (e *HueEmulator) getDevices(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	e.logger.Debug("Devices list is requested")
	response := make(map[string]*Light)

	for _, v := range e.devices {
		response[v.internalHash] = getDevice(v)
	}

	e.sendJSON(w, &LightsList{
		Lights: response,
	})
}

// Responds to GetLightInfo API.
func (e *HueEmulator) getDeviceInfo(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	lightID := params.ByName("lightID")
	for _, v := range e.devices {
		if lightID == v.internalHash {
			e.logger.Debug("Requested device state info", common.LogDeviceNameToken, v.DeviceID)
			e.sendJSON(w, getDevice(v))
			return
		}
	}

	e.logger.Warn("Requested unknown device state info", common.LogDeviceNameToken, lightID)
}

// Updates device state API. Sends command message to master.
func (e *HueEmulator) setDeviceState(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()
	req := &StateCmd{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		e.logger.Error("Failed to decode state change request", err)
		return
	}

	lightID := params.ByName("lightID")
	for _, v := range e.devices {
		if lightID == v.internalHash {

			e.logger.Debug("Requested device command", common.LogDeviceNameToken, v.DeviceID)
			if req.On != nil {
				cmd := enums.CmdOn
				if !*req.On {
					cmd = enums.CmdOff
				}

				e.communicator.Publish(&DeviceCommandMessage{
					Command:    cmd,
					DeviceID:   v.DeviceID,
					Attributes: nil,
				})
			}

			if req.Bri != nil {
				e.communicator.Publish(&DeviceCommandMessage{
					Command:  enums.CmdSetBrightness,
					DeviceID: v.DeviceID,
					Attributes: &common.Percent{
						Value: uint8((float32(*req.Bri) * 100.0) / float32(brightnessMax)),
					},
				})
			}

			m := make(map[string]interface{})
			if req.On != nil {
				m["/lights/"+lightID+"/state/on"] = *req.On
			}
			if req.Bri != nil {
				m["/lights/"+lightID+"/state/bri"] = *req.Bri
			}

			var res [1]struct {
				Success map[string]interface{} `json:"success"`
			}
			res[0].Success = m

			e.sendJSON(w, &res)
			return
		}
	}
}

// Wraps object to JSON.
func (e *HueEmulator) sendJSON(w http.ResponseWriter, val interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		e.logger.Error("Failed to respond", err)
		return
	}
	w.Write(buf.Bytes()) // nolint: gosec
}

// Wraps internal device state into HUE format.
func getDevice(internal *DeviceUpdateMessage) *Light {
	return &Light{
		Type:             "Extended color light",
		ModelID:          "LCT001",
		SWVersion:        "65003148",
		ManufacturerName: "Philips",
		Name:             internal.Name,
		UniqueID:         internal.DeviceID,
		State: State{
			Reachable: true,
			Bri:       getHueBrightness(internal),
			On:        getIsOn(internal),
		},
	}
}

// Transforms internal brightness to HUE format.
func getHueBrightness(internal *DeviceUpdateMessage) uint8 {
	bri, ok := internal.State[enums.PropBrightness]
	if !ok {
		return brightnessMax
	}

	b, err := helpers.UnmarshalProperty(bri, enums.PropBrightness)
	if err != nil {
		return brightnessMax
	}

	return uint8(float32(b.(common.Percent).Value) * float32(brightnessMax) / 100.0)
}

// Transforms internal ON status to HUE format.
func getIsOn(internal *DeviceUpdateMessage) bool {
	on, ok := internal.State[enums.PropOn]
	if !ok {
		return false
	}

	return on.(bool)
}

// Returns simple int hash for the deviceID. This is required since
// emulator-consumers are using same int ID for calling get/set state API.
func hash(s string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return ""
	}
	return fmt.Sprint(h.Sum32())
}
