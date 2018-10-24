package main

import (
	"net"
	"sync"

	"go-home.io/x/server/plugins/api"
	"go-home.io/x/server/plugins/common"
)

// HueEmulator implements extended API plugin and
// provides emulated HUE hub.
type HueEmulator struct {
	sync.Mutex

	Settings *Settings

	logger       common.ILoggerProvider
	isMaster     bool
	communicator api.IExtendedAPICommunicator

	devices            map[string]*DeviceUpdateMessage
	unsupportedDevices []string

	chCommands chan []byte

	upnp     *discoverUPNP
	listener *net.TCPListener
}

// Init starts plugin.
func (e *HueEmulator) Init(data *api.InitDataAPI) error {
	e.communicator = data.Communicator
	e.isMaster = data.IsMaster
	e.logger = data.Logger

	if data.IsMaster {
		return e.initMaster(data)
	}

	return e.initWorker(data)
}

// Routes returns nothing, since we're not exposing anything
// user-related.
func (e *HueEmulator) Routes() []string {
	return []string{}
}

// Unload stops internal processing cycles.
//noinspection GoUnhandledErrorResult
func (e *HueEmulator) Unload() {
	close(e.chCommands)

	if !e.isMaster {
		e.listener.Close() // nolint: gosec
		e.upnp.Stop()
	}
}
