package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	_MYCONTAINER_CONFIGPIPE = "_MYCONTAINER_CONFIGPIPE=3"
	_MYCONTAINER_EXECFIFO   = "_MYCONTAINER_EXECFIFO=4"
	_MYCONTAINER_CONFIGID   = "_MYCONTAINER_CONFIGID="
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

type CgroupConfig struct {
	MemoryLimit int64
	PidLimit    int64
	CpuQuota    int64
	CpuPeriod   int64
}

const (
	PidDefault    = 1024
	PidMinDefault = 16
	PidMaxDefault = 10000000
)

const (
	MemoryDefaultMib = math.MaxInt
	MemoryMinMib     = 16
	MemoryMaxMib     = 1048576
)

const (
	CpuMinPercentage = 20
	CpuMaxPercentage = 100
	CpuQuotaDefault  = 50000
	CpuPeriodDefault = 100000
)

func validate(config *ContainerConfig) error {
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

func writeCgroups(config *CgroupConfig, path *string) error {
	var memory string
	if config.MemoryLimit == MemoryDefaultMib {
		memory = "max"
	} else {
		memBytes := uint64(config.MemoryLimit * 1024 * 1204)
		memory = strconv.FormatUint(memBytes, 10)
	}
	err := os.WriteFile(filepath.Join(*path, "memory.max"), []byte(memory), 0o644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(*path, "pids.max"), []byte(strconv.FormatInt(config.PidLimit, 10)), 0o644)
	if err != nil {
		return err
	}

	cpumax := fmt.Sprintf("%d %d", config.CpuQuota, config.CpuPeriod)
	err = os.WriteFile(filepath.Join(*path, "cpu.max"), []byte(cpumax), 0o644)
	return err
}

func writePidToCgroups(pid int, path *string) error {
	strPid := strconv.Itoa(pid)
	err := os.WriteFile(*path, []byte(strPid), 0o644)
	return err
}

func normalizeCgroup(config *ResourceConfig) CgroupConfig {
	cgroup := CgroupConfig{
		MemoryLimit: MemoryDefaultMib,
		PidLimit:    PidDefault,
		CpuQuota:    CpuQuotaDefault,
		CpuPeriod:   CpuPeriodDefault,
	}
	if config == nil {
		return cgroup
	}

	quota, period := normalizeCPU(config.CPUShares)

	cgroup.MemoryLimit = normalizeMemory(config.MemoryLimit)
	cgroup.PidLimit = normalizePid(config.PidsLimit)
	cgroup.CpuQuota = quota
	cgroup.CpuPeriod = period

	return cgroup
}

func normalizeMemory(memoryConfig int64) int64 {
	switch {
	case memoryConfig > MemoryMaxMib:
		memoryConfig = MemoryDefaultMib
	case memoryConfig < MemoryMinMib:
		memoryConfig = MemoryDefaultMib
	}

	return memoryConfig
}

func normalizePid(pid int64) int64 {
	switch {
	case pid > PidMaxDefault:
		pid = PidDefault
	case pid < PidMinDefault:
		pid = PidDefault
	}

	return pid
}

func normalizeCPU(cpu int64) (int64, int64) {
	var cpuQuota, cpuPercentage int64
	cpuPercentage = CpuPeriodDefault

	switch {
	case cpu > CpuMaxPercentage:
		cpuQuota = CpuQuotaDefault
	case cpu < CpuMinPercentage:
		cpuQuota = CpuQuotaDefault
	}

	if cpuQuota == 0 {
		cpuQuota = (cpu / 100.0) * CpuPeriodDefault
	}

	return cpuQuota, cpuPercentage
}

func create(pathConfig string) (*exec.Cmd, error) {
	// this path should not be in the json config
	// it should be dynamically created the mycontainer part is temporary
	cgroupPath := "/sys/fs/cgroup/mycontainer"
	path := filepath.Join(pathConfig, "config.json")

	jsonConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ContainerConfig

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return nil, err
	}

	if !filepath.IsAbs(config.Rootfs) {
		rootfsPath := filepath.Join(pathConfig, config.Rootfs)
		config.Rootfs = rootfsPath
	}

	err = validate(&config)
	if err != nil {
		return nil, err
	}

	// after validating the pid the full group path is created
	cgroupDir := filepath.Join(cgroupPath, config.ID)
	err = os.Mkdir(cgroupDir, 0o700)
	if err != nil {
		return nil, err
	}

	cgroupConfig := normalizeCgroup(config.Resources)
	err = writeCgroups(&cgroupConfig, &cgroupDir)
	if err != nil {
		return nil, err
	}

	err = writeCgroups(&cgroupConfig, &cgroupDir)
	if err != nil {
		return nil, err
	}

	// create process state
	stateDir := filepath.Join("/run/mycontainer", config.ID)

	err = os.MkdirAll(stateDir, 0o711)
	if err != nil {
		return nil, err
	}

	lockFilePath := filepath.Join(stateDir, "lock")

	fileLock, err := os.OpenFile(lockFilePath, syscall.O_RDWR|syscall.O_CREAT, 0o666)
	if err != nil {
		return nil, err
	}

	defer fileLock.Close()

	// TODO: span a child process investigate exec.fifo is it the child rexec this process for now temp pid 0
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: config.CloneFlags(), Setsid: true}

	cmd.Env = append(cmd.Env, _MYCONTAINER_CONFIGPIPE)

	execPath := filepath.Join(stateDir, "exec.fifo")
	err = syscall.Mkfifo(execPath, 0o622)
	if err != nil {
		return nil, err
	}

	cmd.Env = append(cmd.Env, _MYCONTAINER_CONFIGID+config.ID)

	cmd.Env = append(cmd.Env, _MYCONTAINER_EXECFIFO)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = writePidToCgroups(cmd.Process.Pid, &cgroupDir)
	if err != nil {
		return nil, err
	}

	r.Close()

	state := ContainerState{
		ID:      config.ID,
		PID:     cmd.Process.Pid,
		Status:  "created",
		Bundle:  pathConfig,
		Created: time.Now().UTC(),
		Config:  config,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(configData)
	if err != nil {
		return nil, err
	}

	w.Close()

	stateDirPath := filepath.Join(stateDir, "state.json")
	err = os.WriteFile(stateDirPath, data, 0o644)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
