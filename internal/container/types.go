package main

import "time"

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
