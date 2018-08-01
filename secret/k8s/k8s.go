package main

import (
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/secret"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8SSecretsProvider descries k8s-secret secret store plugin implementation.
type K8SSecretsProvider struct {
	SecretName string
	Namespace  string

	logger       common.ILoggerProvider
	k8sClientSet *kubernetes.Clientset
	secret       *core.Secret
}

// Init makes an attempt to connect to k8s API server and get a secret.
func (s *K8SSecretsProvider) Init(data *secret.InitDataSecret) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		data.Logger.Error("Failed to read k8s secret", err)
		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		data.Logger.Error("Failed to connect to k8s API server", err)
		return err
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

	if err != nil {
		data.Logger.Error("Failed to get secret", err)
		return err
	}

	s.k8sClientSet = clientSet
	s.logger = data.Logger
	s.secret, err = s.k8sClientSet.CoreV1().Secrets(s.Namespace).Get(s.SecretName, v1.GetOptions{})

	if err != nil {
		data.Logger.Error("Failed to get secret", err, "name", s.SecretName, "namespace", s.Namespace)
		return err
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
	sec, err := s.k8sClientSet.CoreV1().Secrets(s.Namespace).Update(s.secret)
	if err != nil {
		s.logger.Error("Failed to update secret", err, "name", s.SecretName, "namespace", s.Namespace)
		return err
	}

	s.secret = sec
	return nil
}

// UpdateLogger performs internal logger update.
// This is necessary since this plugin loads before the system plugin.
func (s *K8SSecretsProvider) UpdateLogger(provider common.ILoggerProvider) {
	s.logger = provider
}
