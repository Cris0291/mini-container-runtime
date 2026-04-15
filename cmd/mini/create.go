package main

import (
	"encoding/json"
	"fmt"
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

type ProcessConfig struct {
	Args []string
	Env  []string
	Cwd  string
	UID  string
	GID  string
}

type Mount struct {
	Source      string
	Destination string
	Type        string
	Flags       int
	Data        string
}

type Namespace struct {
	Type string
	Path string
}

type ResourceConfig struct {
	MemoryLimit int64
	CpuShares   int64
	PidsLimit   int64
}

type NetworkingConfig struct {
	IP      string
	GateWay string
	Bridge  string
}

func create(jsonConfig string) {
	jsonConfigSlice := []byte(jsonConfig)
	err := json.Unmarshal(jsonConfigSlice, &ContainerConfig)
	if err != nil {
		fmt("error")
	}
}
