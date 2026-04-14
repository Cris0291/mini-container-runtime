package main

import (
	"encoding/json"
)

type ContainerConfig struct {
	ID       string
	Hostname string

	Process ProcessConfig
	Rootfs  string
	Mounts  []Mount

	Namespaces []Namespace

	Resources *ResourceConfig

	Networking *NetworkingConfig
}

type ProcessConfig struct{}

func create(jsonConfig string) {
}
