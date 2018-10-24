package main

import (
	"fmt"
	"reflect"
	"time"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
	"go-home.io/x/server/plugins/trigger"
)

// StateTrigger implements events trigger interface.
type StateTrigger struct {
	Settings *Settings

	logger    common.ILoggerProvider
	secret    common.ISecretProvider
	triggered chan interface{}

	deviceUpdates chan *common.MsgDeviceUpdate
}

// TriggerDeviceStateUpdate contains information about device which was the last one
// to trigger this provider and changed property.
type TriggerDeviceStateUpdate struct {
	ID       string
	Property enums.Property
	Value    interface{}
}

// Init starts internal cycle of devices' updates listening.
func (t *StateTrigger) Init(data *trigger.InitDataTrigger) error {
	t.logger = data.Logger
	t.secret = data.Secret
	t.triggered = data.Triggered
	_, t.deviceUpdates = data.FanOut.SubscribeDeviceUpdates()

	go t.internalCycle()
	return nil
}

// Internal cycle of events
func (t *StateTrigger) internalCycle() {
	for msg := range t.deviceUpdates {
		if msg.FirstSeen && !t.Settings.Pessimistic {
			go t.react(msg, true)
			continue
		}

		go t.react(msg, false)
	}
}

// Reacts on device update msg.
func (t *StateTrigger) react(msg *common.MsgDeviceUpdate, noTrigger bool) {
	t.Settings.Lock()
	defer t.Settings.Unlock()

	for _, v := range t.Settings.Devices {
		if !v.deviceRegexp.Match(msg.ID) {
			continue
		}

		for p, s := range msg.State {
			if v.Property != p {
				continue
			}

			v.triggered = t.isTriggered(s, v)
			if !v.triggered {
				continue
			}

			if noTrigger {
				break
			}

			msg := &TriggerDeviceStateUpdate{
				ID:       msg.ID,
				Property: v.Property,
				Value:    v.State,
			}

			t.makeDecision(msg, v)
			break
		}
	}
}

// Checks whether device is triggered.
func (t *StateTrigger) isTriggered(state interface{}, device *DeviceEntry) bool {
	if device.mapperExpr != nil {
		return t.isTriggeredMapper(state, device)
	}

	return t.isTriggeredState(state, device)
}

// Checks whether mapper state triggered.
func (t *StateTrigger) isTriggeredMapper(state interface{}, device *DeviceEntry) bool {
	ok, err := device.mapperExpr.Parse(fmt.Sprintf("%+v", helpers.PlainProperty(state, device.Property)))
	if err != nil {
		t.logger.Error("Failed to execute state trigger mapper", err,
			common.LogDevicePropertyToken, device.Property.String())
		return false
	}

	val, isBool := ok.(bool)
	if !isBool {
		t.logger.Error("Mapper returned non-bool value", err,
			common.LogDevicePropertyToken, device.Property.String())
		return false
	}

	return val
}

// Checks whether state triggered.
func (t *StateTrigger) isTriggeredState(state interface{}, device *DeviceEntry) bool {
	return reflect.DeepEqual(state, device.State)
}

// Makes a decision whether need to trigger.
func (t *StateTrigger) makeDecision(msg *TriggerDeviceStateUpdate, spec *DeviceEntry) {
	if t.Settings.decisionLogic == logicOr {
		go t.triggerOr(msg, spec)
		return
	}

	if t.Settings.decisionLogic == logicAnd {
		for _, v := range t.Settings.Devices {
			if !v.triggered {
				return
			}
		}

		go t.triggerAnd(msg)
	}
}

// Checks conditions if trigger's logic is "OR".
func (t *StateTrigger) triggerOr(msg *TriggerDeviceStateUpdate, spec *DeviceEntry) {
	if t.Settings.Delay > 0 {
		time.Sleep(time.Duration(t.Settings.Delay) * time.Second)
		if !spec.triggered {
			return
		}
	}

	t.triggered <- msg
	t.logger.Info("Triggering due to OR logic", common.LogDevicePropertyToken, msg.Property.String(),
		logTokenTargetDevice, msg.ID)
}

// Checks conditions if trigger's logic is "AND".
func (t *StateTrigger) triggerAnd(msg *TriggerDeviceStateUpdate) {
	if t.Settings.Delay > 0 {
		time.Sleep(time.Duration(t.Settings.Delay) * time.Second)
		t.Settings.Lock()
		defer t.Settings.Unlock()
		for _, v := range t.Settings.Devices {
			if !v.triggered {
				return
			}
		}
	}

	t.triggered <- msg
	t.logger.Info("Triggering due to AND logic", common.LogDevicePropertyToken, msg.Property.String(),
		logTokenTargetDevice, msg.ID)
}
