// Package main contains k8s-secret implementation for the go-home secrets storage.
package main

// Load is the main plugin entry point.
// nolint: deadcode
func Load() (interface{}, interface{}, error) {
	return &K8SSecretsProvider{}, nil, nil
}
