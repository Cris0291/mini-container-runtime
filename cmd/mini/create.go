package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"syscall"
)

type ContainerConfig struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`

	Process ProcessConfig `json:"process_config"`
	Rootfs  string        `json:"rootfs"`
	Mounts  []Mount       `json:"mounts"`

	Namespaces []Namespace `json:"namespaces"`

	Resources *ResourceConfig `json:"resources"`

	Networking *NetworkingConfig `json:"networking"`
}

//TODO:rethink map definition on the create global

namespaceRelation := map[string][int]{

}

func (c *ContainerConfig) CloneFlags() uintptr{
	for namespace := range c.Namespaces {

	}
}

type ProcessConfig struct {
	Args []string `json:"args"`
	Env  []string `json:"env"`
	Cwd  string   `json:"cwd"`
	UID  int      `json:"uid"`
	GID  int      `json:"gid"`
}

type Mount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Type        string `json:"type"`
	Flags       int    `json:"flags"`
	Data        string `json:"data"`
}

type Namespace struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type ResourceConfig struct {
	MemoryLimit int64 `json:"memory_limit"`
	CPUShares   int64 `json:"cpu_shares"`
	PidsLimit   int64 `json:"pids_limit"`
}

type NetworkingConfig struct {
	IP      string `json:"ip"`
	GateWay string `json:"gateway"`
	Bridge  string `json:"bridge"`
}

type ContainerState struct {
	ID      string
	PID     int
	Status  string
	Bundle  string
	Created time.Time
	Config  ContainerConfig
}

func Validate(config *ContainerConfig) error {
	if config.ID == "" {
		return errors.New("No id was provided in the json file")
	}
	if config.Hostname == "" {
		return errors.New("NO hostname was provided i the json config file")
	}
	if config.Rootfs == "" {
		return errors.New("No rootfs was provided in the json config file")
	}
	return nil
}

func create(pathConfig string) error {
	path := filepath.Join(pathConfig, "config.json")

	jsonConfig, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config ContainerConfig

	e := json.Unmarshal(jsonConfig, &config)
	Validate(&config)
	if e != nil {
		return e
	}

	// create process state
	stateDir := filepath.Join("/run/mycontainer", config.ID)

	er := os.MkdirAll(stateDir, 0o711)
	if er != nil {
		return er
	}

	// TODO: span a child process investigate exec.fifo is it the child rexec this process for now temp pid 0

	state := ContainerState{
		ID:      config.ID,
		PID:     0,
		Status:  "created",
		Bundle:  pathConfig,
		Created: time.Now().UTC(),
		Config:  config,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	stateDirPath := filepath.Join(stateDir, "state.json")
	os.WriteFile(stateDirPath, data, 0o644)

	return nil
}
