package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
	PID  string
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
	CPUShares   int64
	PidsLimit   int64
}

type NetworkingConfig struct {
	IP      string
	GateWay string
	Bridge  string
}

type ContainerState struct {
	ID      string
	PID     string
	Status  string
	Bundle  string
	Created time.Time
	Config  ContainerConfig
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

	var config ContainerConfig

	e := json.Unmarshal(jsonConfig, &config)
	Validate(&config)
	if e != nil {
		fmt.Println("error")
	}

	// create process state
	stateDir := filepath.Join("/run/mycontainer", config.ID)
	er := os.MkdirAll(stateDir, 0o711)

	state := ContainerState{
		ID:      config.ID,
		PID:     config.Process.PID,
		Status:  "created",
		Bundle:  pathConfig,
		Created: time.Now().UTC(),
		Config:  config,
	}
}
