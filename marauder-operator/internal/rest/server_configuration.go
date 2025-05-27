package rest

import (
	"fmt"

	"github.com/docker/docker/api/types/registry"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-operator/pkg/manager"
)

// The ServerConfiguration struct holds relevant configuration values for the rest server.
type ServerConfiguration struct {
	Identifier string `yaml:"identifier"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	Controller Controller `yaml:"controller"`

	Docker Docker `yaml:"docker"`

	Disk Disk `yaml:"disk"`

	TLS utils.TLSConfiguration `yaml:"tls"`
}

// Disk contains configuration values for the disk setup of controller.
type Disk struct {
	DownloadPath           string               `yaml:"downloadPath"`
	ServerDataPathTemplate string               `yaml:"serverDataPathTemplate"`
	FolderOwner            *manager.FolderOwner `yaml:"folderOwner,omitempty"`
}

// The Controller struct holds the configuration values for the controller client used by the operator.
type Controller struct {
	Endpoint    string `yaml:"endpoint"`
	WorkerCount int    `yaml:"workerCount"`
}

// Docker represents the docker configuration of the controller.
type Docker struct {
	Username              string `yaml:"username"`
	Password              string `yaml:"password"`
	AutoRemoveContainers  bool   `yaml:"autoRemoveContainers"`
	ContainerMemoryBuffer int64  `yaml:"containerMemoryBuffer"`
}

// ToBasicAuth converts the docker config into the encoded auth string.
func (d Docker) ToBasicAuth() (string, error) {
	config, err := registry.EncodeAuthConfig(registry.AuthConfig{Username: d.Username, Password: d.Password})
	if err != nil {
		return "", fmt.Errorf("failed to encode auth config: %w", err)
	}

	return config, nil
}
