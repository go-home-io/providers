// Package main contains Zengge lights implementation for the go-home light.
// Those devices are also known as Flux or Magic Home.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	settings := &Settings{}

	return &ZenggeLight{Settings: settings}, settings, nil
}
