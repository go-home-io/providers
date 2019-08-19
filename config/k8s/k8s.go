package main

import (
	"context"
	"strings"

	"github.com/ericchiang/k8s"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/config"
)

// K8SConfigProvider descries k8s-config-map configs plugin implementation.
type K8SConfigProvider struct {
	ConfigMapName string
	Namespace     string

	logger    common.ILoggerProvider
	k8sClient *k8s.Client
}

// Init makes an attempt to connect to k8s API server and get a config-map.
func (c *K8SConfigProvider) Init(data *config.InitDataConfig) error {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		data.Logger.Error("Failed to connect to k8s API server", err)
		return errors.Wrap(err, "k8s is not available")
	}

	cm, ok := data.Options["config-map"]

	if !ok {
		data.Logger.Warn("Config map name is not provided, using 'go-home'")
		cm = "go-home"
	}

	parts := strings.Split(cm, "/")
	if len(parts) < 2 {
		data.Logger.Warn("Config map namespace is not provided, using 'default'")
		c.Namespace = "default"
		c.ConfigMapName = parts[0]
	} else {
		c.Namespace = parts[0]
		c.ConfigMapName = parts[1]
	}

	c.k8sClient = client
	c.logger = data.Logger
	return nil
}

// Load makes an attempt to read stings data from a k8s config map.
func (c *K8SConfigProvider) Load() chan []byte {
	var cm v1.ConfigMap

	err := c.k8sClient.Get(context.Background(), c.Namespace, c.ConfigMapName, &cm)
	if err != nil {
		c.logger.Error("Failed to get config map", err,
			"config-map", c.ConfigMapName, "namespace", c.Namespace)
		return nil
	}
	dataChan := make(chan []byte)

	go func() {
		for k, v := range cm.Data {
			if !config.IsValidConfigFileName(k) {
				continue
			}

			c.logger.Info("Processing config map entry", "name", k,
				"config-map", c.ConfigMapName, "namespace", c.Namespace)
			dataChan <- []byte(v)
		}

		close(dataChan)
	}()

	return dataChan
}
