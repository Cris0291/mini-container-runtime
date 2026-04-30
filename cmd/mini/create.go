package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var _MYCONTAINER_CONFIGPIPE = "3"

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

// TODO:rethink map definition on the create global
var namespaceRelation = map[string]uintptr{
	"pid":    syscall.CLONE_NEWPID,
	"uts":    syscall.CLONE_NEWUTS,
	"mount":  syscall.CLONE_NEWNS,
	"net":    syscall.CLONE_NEWNET,
	"ipc":    syscall.CLONE_NEWIPC,
	"user":   syscall.CLONE_NEWUSER,
	"cgroup": syscall.CLONE_NEWCGROUP,
}

func (c *ContainerConfig) CloneFlags() uintptr {
	var flags uintptr
	for _, namespace := range c.Namespaces {
		if strings.TrimSpace(namespace.Path) == "" {
			value, ok := namespaceRelation[namespace.Type]
			if ok {
				flags |= value
			}
		}
	}
	return flags
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
	ID      string          `json:"id"`
	PID     int             `json:"pid"`
	Status  string          `json:"status"`
	Bundle  string          `json:"bundle"`
	Created time.Time       `json:"created"`
	Config  ContainerConfig `json:"container_config"`
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
	_, err := os.Stat(config.Rootfs)
	if err != nil {
		return errors.New("Rootfs path does not exist")
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

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(config.Rootfs) {
		rootfsPath := filepath.Join(pathConfig, config.Rootfs)
		config.Rootfs = rootfsPath
	}

	err = Validate(&config)
	if err != nil {
		return err
	}

	// create process state
	stateDir := filepath.Join("/run/mycontainer", config.ID)

	err = os.MkdirAll(stateDir, 0o711)
	if err != nil {
		return err
	}

	// TODO: span a child process investigate exec.fifo is it the child rexec this process for now temp pid 0
	r, w, err := os.Pipe()

	cmd := exec.Command("proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)

	cmd.SysProcAttr.Cloneflags = config.CloneFlags()

	cmd.ExtraFiles = append(cmd.ExtraFiles, r)

	cmd.Env = append(cmd.Env, _MYCONTAINER_CONFIGPIPE)

	err = cmd.Start()
	if err != nil {
		return err
	}

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
	err = os.WriteFile(stateDirPath, data, 0o644)
	if err != nil {
		return err
	}

	return nil
}
