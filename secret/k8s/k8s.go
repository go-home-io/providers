package main

import (
	"context"
	"strings"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/apis/core/v1"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/secret"
	"github.com/pkg/errors"
)

// K8SSecretsProvider descries k8s-secret secret store plugin implementation.
type K8SSecretsProvider struct {
	SecretName string
	Namespace  string

	logger    common.ILoggerProvider
	k8sClient *k8s.Client
	secret    v1.Secret
}

// Init makes an attempt to connect to k8s API server and get a secret.
func (s *K8SSecretsProvider) Init(data *secret.InitDataSecret) error {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		data.Logger.Error("Failed to connect to k8s API server", err)
		return errors.Wrap(err, "k8s is not available")
	}

	sec, ok := data.Options["secret"]

	if !ok {
		data.Logger.Warn("Secret name is not provided, using 'go-home'")
		sec = "go-home"
	}

	parts := strings.Split(sec, "/")
	if len(parts) < 2 {
		data.Logger.Warn("Secret namespace is not provided, using 'default'")
		s.Namespace = "default"
		s.SecretName = parts[0]
	} else {
		s.Namespace = parts[0]
		s.SecretName = parts[1]
	}

	s.k8sClient = client
	s.logger = data.Logger

	err = s.k8sClient.Get(context.Background(), s.Namespace, s.SecretName, &s.secret)
	if err != nil {
		data.Logger.Error("Failed to get secret", err, "name", s.SecretName, "namespace", s.Namespace)
		return errors.Wrap(err, "secret get failed")
	}

	return nil
}

// Get performs an attempt to get secret value.
func (s *K8SSecretsProvider) Get(name string) (string, error) {
	data, ok := s.secret.Data[name]
	if !ok {
		return "", errors.New("not found")
	}

	return string(data), nil
}

// Set performs an attempt to update k8s secret.
func (s *K8SSecretsProvider) Set(name string, data string) error {
	s.secret.Data[name] = []byte(data)
	err := s.k8sClient.Update(context.Background(), &s.secret)
	if err != nil {
		s.logger.Error("Failed to update secret", err, "name", s.SecretName, "namespace", s.Namespace)
		return errors.Wrap(err, "secret update failed")
	}

	return nil
}

// UpdateLogger performs internal logger update.
// This is necessary since this plugin loads before the system plugin.
func (s *K8SSecretsProvider) UpdateLogger(provider common.ILoggerProvider) {
	s.logger = provider
}
