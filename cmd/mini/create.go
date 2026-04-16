package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func Validate(config *ContainerConfig) {
	if config.ID == "" {
		fmt.Println("empty id in container creation")
	}
}

func create(pathConfig string) {
	path := filepath.Join(pathConfig, "config.json")

	jsonConfig, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read config error", err)
	}

	config := new(ContainerConfig)

	e := json.Unmarshal(jsonConfig, &config)
	Validate(config)
	if e != nil {
		fmt.Println("error")
	}
}
