// Package main contains k8s-config-map implementation for the go-home configs storage.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	return &K8SConfigProvider{}, nil, nil
}
