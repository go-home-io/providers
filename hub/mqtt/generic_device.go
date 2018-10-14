package main

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/pkg/errors"
)

// Mapper for command with corresponding topic and
// template expression.
type commandMapper struct {
	topic      string
	expression helpers.ITemplateExpression
}

// Mapper for properties.
type propertyMapper struct {
	commandMapper
	property enums.Property
}

// IGenericDevice defines reconnect feature for the device.
type IGenericDevice interface {
	ReConnect()
}

// Generic MQTT device
type mqttDevice struct {
	mutex sync.Mutex

	topicsPrefix string
	state        interface{}
	updateChan   chan *device.StateUpdateData
	topics       map[string]*propertyMapper
	settings     *DeviceSettings
	spec         *device.Spec
	parser       helpers.ITemplateParser
	logger       common.ILoggerProvider
	client       mqtt.Client
	commands     map[enums.Command]*commandMapper
	uom          enums.UOM
}

// Handles update from MQTT broker.
func (m *mqttDevice) handleUpdates(payload []byte, expression helpers.ITemplateExpression, property enums.Property) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	msg := string(payload)
	val, err := expression.Parse(msg)

	if err != nil {
		m.logger.Error("Failed to map received mqtt property", err,
			common.LogDevicePropertyToken, property.String(), "message", msg)
		return
	}

	propName := property.GetPropertyName()

	rt, rv := reflect.TypeOf(m.state), reflect.ValueOf(m.state)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	_, ok := rt.FieldByName(propName)
	if !ok {
		m.logger.Error("Failed to get mqtt property", err, common.LogDevicePropertyToken, property.String())
		return
	}

	fieldV := rv.FieldByName(propName)
	if fieldV.Kind() == reflect.Ptr {
		fieldV = fieldV.Elem()
	}

	val = helpers.PropertyFixNum(val, property)

	val = helpers.UOMConvertInterface(val, property, m.settings.UOM, m.uom)
	if helpers.PropertyDeepEqual(fieldV.Interface(), val, property) {
		return
	}

	fieldV.Set(reflect.ValueOf(val))
	m.logger.Debug("Received update for mqtt device", common.LogDevicePropertyToken, property.String())

	// It' possible to receive update before actual load
	if nil == m.updateChan {
		return
	}

	m.forceUpdate()
}

// Re-connecting to broker.
func (m *mqttDevice) ReConnect() {
	for _, v := range m.topics {
		topic := v.topic
		expr := v.expression
		prop := v.property
		m.logger.Debug("Re-subscribing to MQTT", logTokenTopic, topic)
		m.client.Unsubscribe(topic)
		m.client.Subscribe(topic, m.settings.Qos, func(client mqtt.Client, message mqtt.Message) {
			go m.handleUpdates(message.Payload(), expr, prop)
		})
	}
}

// Subscribes to an MQTT topic.
func (m *mqttDevice) subscribe() {
	m.topics = make(map[string]*propertyMapper)
	m.commands = make(map[enums.Command]*commandMapper)
	m.spec = &device.Spec{
		SupportedCommands:   make([]enums.Command, 0),
		SupportedProperties: make([]enums.Property, 0),
	}

	if m.settings.Type == enums.DevSensor {
		m.spec.SupportedProperties = append(m.spec.SupportedProperties, enums.PropSensorType)
		m.state.(*device.SensorState).SensorType = m.settings.SensorType
	}

	for _, v := range m.settings.Properties {
		topic := fmt.Sprintf("%s%s", m.topicsPrefix, v.Topic)
		expr, err := m.parser.Compile(v.Mapper)
		if err != nil {
			m.logger.Error("Failed to compile mqtt property mapper", err,
				"mapper", v.Mapper, common.LogDevicePropertyToken, v.Property.String())
			continue
		}
		prop := v.Property
		m.logger.Debug("Subscribing to MQTT", logTokenTopic, topic)
		m.client.Subscribe(topic, m.settings.Qos, func(client mqtt.Client, message mqtt.Message) {
			go m.handleUpdates(message.Payload(), expr, prop)
		})

		m.topics[topic] = &propertyMapper{
			commandMapper: commandMapper{
				topic:      topic,
				expression: expr,
			},
			property: v.Property,
		}

		m.spec.SupportedProperties = append(m.spec.SupportedProperties, v.Property)
	}

	for _, v := range m.settings.Commands {
		expr, err := m.parser.Compile(v.Mapper)
		if err != nil {
			m.logger.Error("Failed to compile mqtt command mapper", err,
				"mapper", v.Mapper, common.LogDeviceCommandToken, v.Command.String())
			continue
		}

		cmdMapper := &commandMapper{
			topic:      fmt.Sprintf("%s%s", m.topicsPrefix, v.Topic),
			expression: expr,
		}

		m.commands[v.Command] = cmdMapper
		m.spec.SupportedCommands = append(m.spec.SupportedCommands, v.Command)
	}
}

// Handles command invocation.
func (m *mqttDevice) command(cmd enums.Command) error {
	cmdMap := m.commands[cmd]
	val, err := cmdMap.expression.Format(map[string]interface{}{"state": m.state})
	if err != nil {
		m.logger.Error("Failed to format received mqtt property", err,
			common.LogDeviceCommandToken, cmd.String())
		return errors.Wrap(err, "command failed")
	}

	m.client.Publish(cmdMap.topic, m.settings.Qos, m.settings.Retained, val)

	return nil
}

// Forces pushing an update message.
func (m *mqttDevice) forceUpdate() {
	if nil != m.updateChan {
		m.updateChan <- &device.StateUpdateData{
			State: m.state,
		}
	}
}

// Init performs initial plugin initialization.
func (m *mqttDevice) Init(data *device.InitDataDevice) error {
	m.updateChan = data.DeviceStateUpdateChan
	m.logger = data.Logger
	return nil
}

// Unload handles un-subscribing from MQTT broker.
func (m *mqttDevice) Unload() {
	for _, v := range m.topics {
		m.client.Unsubscribe(v.topic)
	}
}

// GetName returns device name.
func (m *mqttDevice) GetName() string {
	return m.settings.Name
}

// GetSpec returns device spec.
func (m *mqttDevice) GetSpec() *device.Spec {
	return m.spec
}
