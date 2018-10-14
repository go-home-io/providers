package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
)

const (
	// Describes internal nsq log level.
	nsqLogLevel = nsq.LogLevelWarning
)

// NsqBus describes NSQ bus plugin implementation.
type NsqBus struct {
	sync.Mutex

	config    *nsq.Config
	producer  *nsq.Producer
	logger    common.ILoggerProvider
	Settings  *Settings
	consumers map[string]*nsq.Consumer
}

// Init makes an attempt to setup a new NSQ producer.
func (b *NsqBus) Init(data *bus.InitDataServiceBus) error {
	b.config = nsq.NewConfig()
	b.config.ClientID = data.NodeID
	b.config.DialTimeout = time.Duration(b.Settings.Timeout) * time.Second
	b.config.LookupdPollInterval = 5 * time.Second

	var err error
	b.producer, err = nsq.NewProducer(b.Settings.ServerAddress, b.config)
	if err != nil {
		return errors.Wrap(err, "producer init failed")
	}

	b.logger = data.Logger
	b.producer.SetLogger(&nsqLogger{logger: data.Logger}, nsqLogLevel)

	err = b.producer.Ping()
	if err != nil {
		return errors.Wrap(err, "ping failed")
	}

	return nil
}

// Subscribe makes an attempts to subscribe to NSQ topic.
func (b *NsqBus) Subscribe(channel string, queue chan bus.RawMessage) error {
	b.Lock()
	defer b.Unlock()

	if _, ok := b.consumers[channel]; ok {
		b.logger.Warn("Trying to subscribe to the same channel twice",
			common.LogChannelToken, channel)
		return nil
	}

	q, err := nsq.NewConsumer(channel, "gh", b.config)
	if err != nil {
		b.logger.Error("Failed to subscribe to the channel", err, common.LogChannelToken, channel)
		return errors.Wrap(err, "consumer init failed")
	}

	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		b.Lock()
		defer b.Unlock()
		_, ok := b.consumers[channel]
		if !ok {
			return nil
		}

		msg := bus.RawMessage{
			Body: make([]byte, len(message.Body)),
		}
		copy(msg.Body, message.Body)
		queue <- msg
		return nil
	}))

	q.SetLogger(&nsqLogger{logger: b.logger}, nsqLogLevel)

	err = q.ConnectToNSQD(b.Settings.ServerAddress)
	if err != nil {
		b.logger.Error("Failed to connect to nsq while subscribing", err, common.LogChannelToken, channel)
		return errors.Wrap(err, "connection to bus failed")
	}

	b.logger.Debug("Successfully subscribed to nsq channel", common.LogChannelToken, channel)
	b.consumers[channel] = q

	return nil
}

// Unsubscribe removes channel subscription.
func (b *NsqBus) Unsubscribe(channel string) {
	b.Lock()
	defer b.Unlock()

	if ch, ok := b.consumers[channel]; !ok {
		b.logger.Warn("Trying to unsubscribe from the channel without been subscribed",
			common.LogChannelToken, channel)
	} else {
		ch.Stop()
		delete(b.consumers, channel)
	}
}

// Publish makes an attempt to publish a new message.
func (b *NsqBus) Publish(channel string, messages ...interface{}) {
	for _, m := range messages {
		data, err := json.Marshal(m)
		if err != nil {
			b.logger.Error("Failed to marshal message to channel ", err, common.LogChannelToken, channel)
		}

		err = b.producer.Publish(channel, data)
		if err != nil {
			b.logger.Error("Failed to marshal message to channel", err,
				"msg", string(data), common.LogChannelToken, channel)
		}
	}
}

// Ping validates whether NSQ is available.
func (b *NsqBus) Ping() error {
	err := b.producer.Ping()
	if err != nil {
		b.logger.Error("Service bus is down", err)
		return errors.Wrap(err, "ping failed")
	}

	return nil
}
