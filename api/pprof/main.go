// Package main contains pprof API extension.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	return &PprofAPI{}, nil, nil
}
