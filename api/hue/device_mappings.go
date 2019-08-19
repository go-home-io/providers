package main

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// Transforms device-specific property into ON/OFF status.
func getIsOnDeviceSpecific(internal *DeviceUpdateMessage) bool {
	switch internal.DeviceType { // nolint: gocritic
	case enums.DevVacuum:
		st, ok := internal.State[enums.PropVacStatus]
		if !ok {
			return false
		}

		return st == enums.VacCleaning
	}

	return false
}

// Transforms device-specific property into Brightness status.
func getBrightnessDeviceSpecific(internal *DeviceUpdateMessage) uint8 {
	switch internal.DeviceType { // nolint: gocritic
	case enums.DevVacuum:
		fan, ok := internal.State[enums.PropFanSpeed]
		if !ok {
			return brightnessMax
		}

		b, err := helpers.UnmarshalProperty(fan, enums.PropFanSpeed)
		if err != nil {
			return brightnessMax
		}

		return convertPercentToHueBrightness(b.(common.Percent).Value)
	}

	return brightnessMax
}

// Converts percent into HUE brightness.
func convertPercentToHueBrightness(value uint8) uint8 {
	return uint8(float32(value) * float32(brightnessMax) / 100.0)
}

// Transforms received Brightness value into device command.
func setBrightnessDeviceSpecific(internal *DeviceUpdateMessage, percent uint8) *DeviceCommandMessage {
	switch internal.DeviceType {
	case enums.DevVacuum:
		return &DeviceCommandMessage{
			Command:  enums.CmdSetFanSpeed,
			DeviceID: internal.DeviceID,
			Attributes: &common.Percent{
				Value: percent,
			},
		}
	case enums.DevSwitch:
		return nil
	}

	return &DeviceCommandMessage{
		Command:  enums.CmdSetBrightness,
		DeviceID: internal.DeviceID,
		Attributes: &common.Percent{
			Value: percent,
		},
	}
}
