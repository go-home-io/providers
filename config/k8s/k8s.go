package main

import (
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8SConfigProvider descries k8s-config-map configs plugin implementation.
type K8SConfigProvider struct {
	ConfigMapName string
	Namespace     string

	logger       common.ILoggerProvider
	k8sClientSet *kubernetes.Clientset
}

// Init makes an attempt to connect to k8s API server and get a config-map.
func (c *K8SConfigProvider) Init(data *config.InitDataConfig) error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		data.Logger.Error("Failed to read k8s config", err)
		return err
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		data.Logger.Error("Failed to connect to k8s API server", err)
		return err
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

	c.k8sClientSet = clientSet
	c.logger = data.Logger
	return nil
}

// Load makes an attempt to read stings data from a k8s config map.
func (c *K8SConfigProvider) Load() chan []byte {
	cm, err := c.k8sClientSet.CoreV1().ConfigMaps(c.Namespace).Get(c.ConfigMapName, v1.GetOptions{})
	if err != nil {
		c.logger.Error("Failed to get config map", err, "config-map", c.ConfigMapName, "namespace", c.Namespace)
		return nil
	}
	dataChan := make(chan []byte)

	go func() {
		for k, v := range cm.Data {
			if !config.IsValidConfigFileName(k) {
				continue
			}

			c.logger.Info("Processing config map entry", "name", k, "config-map", c.ConfigMapName, "namespace", c.Namespace)
			dataChan <- []byte(v)
		}

		close(dataChan)
	}()

	return dataChan
}
