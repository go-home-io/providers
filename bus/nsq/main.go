// Package main contains NSQ implementation for the go-home hub.
package main

import (
	"github.com/nsqio/go-nsq"
)

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &NsqBus{
			consumers: make(map[string]*nsq.Consumer),
			Settings:  settings,
		},
		settings, nil
}
