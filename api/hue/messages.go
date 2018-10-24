package main

import (
	"go-home.io/x/server/plugins/api"
	"go-home.io/x/server/plugins/device/enums"
)

// DeviceUpdateMessage has data about device update.
// This message is produced by master.
type DeviceUpdateMessage struct {
	api.ExtendedAPIMessage
	Name       string                         `json:"n"`
	DeviceType enums.DeviceType               `json:"t"`
	DeviceID   string                         `json:"i"`
	State      map[enums.Property]interface{} `json:"s"`

	internalHash string
}

// DeviceCommandMessage has data with new device command.
// This message is produced by worker.
type DeviceCommandMessage struct {
	api.ExtendedAPIMessage
	IsDiscovery bool          `json:"d"`
	DeviceID    string        `json:"i"`
	Command     enums.Command `json:"c"`
	Attributes  interface{}   `json:"a"`
}

// State describes HUE light state.
type State struct {
	Hue       int        `json:"hue"`
	Sat       int        `json:"sat"`
	CT        int        `json:"ct"`
	XY        [2]float64 `json:"xy"`
	Reachable bool       `json:"reachable"`
	On        bool       `json:"on"`
	Bri       uint8      `json:"bri"`
	Effect    string     `json:"effect"`
	Alert     string     `json:"alert"`
	ColorMode string     `json:"colormode"`
}

// Light describes HUE light.
type Light struct {
	State            State  `json:"state"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	ModelID          string `json:"modelid"`
	ManufacturerName string `json:"manufacturername"`
	UniqueID         string `json:"uniqueid"`
	SWVersion        string `json:"swversion"`
	PointSymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

// LightsList describes list of light known by HUE hub.
type LightsList struct {
	Lights map[string]*Light `json:"lights"`
}

// StateCmd describes light command sent to HUE hub.
type StateCmd struct {
	On  *bool `json:"on"`
	Bri *int  `json:"bri"`
}
