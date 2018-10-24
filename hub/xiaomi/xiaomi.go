package main

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/vkorn/go-miio"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
)

// XiaomiHub defines hub device with enabled development API.
type XiaomiHub struct {
	sync.Mutex

	Settings *Settings

	logger        common.ILoggerProvider
	gateway       *miio.Gateway
	discoveryChan chan *device.DiscoveredDevices
	updatesChan   chan *device.StateUpdateData
	stopChan      chan bool

	devices map[string]IXiaomiDevice
}

// Init performs initial device init.
func (h *XiaomiHub) Init(data *device.InitDataDevice) error {
	miio.LOGGER = &xiaomiLogger{
		logger: data.Logger,
	}

	h.discoveryChan = data.DeviceDiscoveredChan
	h.updatesChan = data.DeviceStateUpdateChan
	h.logger = data.Logger
	h.stopChan = make(chan bool, 1)
	h.devices = make(map[string]IXiaomiDevice)

	g, err := miio.NewGateway(h.Settings.IP, h.Settings.Key)
	if err != nil {
		return errors.Wrap(err, "gateway init failed")
	}

	h.gateway = g
	go h.processUpdates()
	return nil
}

// Unload stops multi-cast and device connection listeners.
func (h *XiaomiHub) Unload() {
	h.stopChan <- true
	h.gateway.Stop()
}

// GetName returns hub's IP.
func (h *XiaomiHub) GetName() string {
	return h.Settings.IP
}

// GetSpec returns device spec.
func (h *XiaomiHub) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedProperties: []enums.Property{enums.PropNumDevices},
		SupportedCommands:   []enums.Command{},
	}
}

// Load performs initial load.
func (h *XiaomiHub) Load() (*device.HubLoadResult, error) {
	h.Lock()
	defer h.Unlock()

	return &device.HubLoadResult{
		State: &device.HubState{
			NumDevices: len(h.devices),
		},
	}, nil
}

// Update is not used since device pushes updates.
func (h *XiaomiHub) Update() (*device.HubLoadResult, error) {
	return nil, nil
}

// Processes updates coming from the hub.
func (h *XiaomiHub) processUpdates() {
	for {
		select {
		case stop, ok := <-h.stopChan:
			if stop || !ok {
				return
			}
		case state, ok := <-h.gateway.UpdateChan:
			if !ok {
				return
			}
			go h.processDeviceState(state)
		}
	}
}

// Processes state update messages coming from the hub.
func (h *XiaomiHub) processDeviceState(state *miio.DeviceUpdateMessage) {
	h.Lock()
	defer h.Unlock()

	switch v := state.State.(type) {
	case *miio.GatewayState:
		h.processGateway(state.ID, v)
	case *miio.SwitchState:
		h.processButton(state.ID, v)
	case *miio.SensorHTState:
		h.processHTSensor(state.ID, v)
	case *miio.MagnetState:
		h.processMagnet(state.ID, v)
	case *miio.MotionState:
		h.processMotion(state.ID, v)
	}
}

// Processes gateway state updates.
// nolint: dupl
func (h *XiaomiHub) processGateway(gatewayID string, state *miio.GatewayState) {
	g, ok := h.devices[gatewayID]
	if !ok {
		g = &xiaomiGateway{
			xiaomiDevice: xiaomiDevice{
				deviceID:   gatewayID,
				deviceType: gateway,
			},
			gateway: h.gateway,
			state:   &device.LightState{},
		}

		h.discovery(g, state, enums.DevLight)
	} else {
		g.InternalUpdate(state, false)
	}
}

// Processes button state updates.
// nolint: dupl
func (h *XiaomiHub) processButton(buttonID string, state *miio.SwitchState) {
	b, ok := h.devices[buttonID]
	if !ok {
		b = &xiaomiButton{
			xiaomiDevice: xiaomiDevice{
				deviceID:   buttonID,
				deviceType: button,
			},
			state: &device.SensorState{
				SensorType: enums.SenButton,
			},
		}

		h.discovery(b, state, enums.DevSensor)
	} else {
		b.InternalUpdate(state, false)
	}
}

// Processes temperature/humidity sensor state updates.
func (h *XiaomiHub) processHTSensor(sensorID string, state *miio.SensorHTState) {
	s, ok := h.devices[sensorID]
	if !ok {
		s = &xiaomiTemperatureSensor{
			xiaomiDevice: xiaomiDevice{
				deviceID:   sensorID,
				deviceType: temperature,
			},
			currentUOM: h.Settings.UOM,
			state: &device.SensorState{
				SensorType: enums.SenTemperature,
			},
		}

		h.discovery(s, state, enums.DevSensor)
	} else {
		s.InternalUpdate(state, false)
	}
}

// Processes magnet state updates.
// nolint: dupl
func (h *XiaomiHub) processMagnet(magnetID string, state *miio.MagnetState) {
	s, ok := h.devices[magnetID]
	if !ok {
		s = &xiaomiMagnet{
			xiaomiDevice: xiaomiDevice{
				deviceID:   magnetID,
				deviceType: magnet,
			},
			state: &device.SensorState{
				SensorType: enums.SenLock,
			},
		}

		h.discovery(s, state, enums.DevSensor)
	} else {
		s.InternalUpdate(state, false)
	}
}

// Processes motion state updates.
// nolint: dupl
func (h *XiaomiHub) processMotion(motionID string, state *miio.MotionState) {
	s, ok := h.devices[motionID]
	if !ok {
		s = &xiaomiMotion{
			xiaomiDevice: xiaomiDevice{
				deviceID:   motionID,
				deviceType: motion,
			},
			state: &device.SensorState{
				SensorType: enums.SenMotion,
			},
		}

		h.discovery(s, state, enums.DevSensor)
	} else {
		s.InternalUpdate(state, false)
	}
}

// Sends a new discovery message to the worker.
// nolint: dupl
func (h *XiaomiHub) discovery(d IXiaomiDevice, state interface{}, deviceType enums.DeviceType) {
	h.devices[d.GetID()] = d
	s := d.InternalUpdate(state, true)
	h.discoveryChan <- &device.DiscoveredDevices{
		State:     s,
		Type:      deviceType,
		Interface: d,
	}

	h.updatesChan <- &device.StateUpdateData{
		State: &device.HubState{
			NumDevices: len(h.devices),
		},
	}
}
