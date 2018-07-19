package main

import (
	"reflect"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/trigger"
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
			continue
		}

		go t.react(msg)
	}
}

// Reacts on device update msg.
func (t *StateTrigger) react(msg *common.MsgDeviceUpdate) {
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

			v.triggered = reflect.DeepEqual(s, v.State)
			if !v.triggered {
				continue
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
		common.LogDeviceNameToken, msg.ID)
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
		common.LogDeviceNameToken, msg.ID)
}
