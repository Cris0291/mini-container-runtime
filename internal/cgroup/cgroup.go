package main

type CgroupContainer struct {
	Config         CgroupConfig
	Path           string
	SubControlPath string
	GroupsPath     string
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
	MemoryDefaultMib = 0
	MemoryMinMib     = 16
	MemoryMaxMib     = 1048576
)

const (
	CpuMinPercentage = 20
	CpuMaxPercentage = 100
	CpuQuotaDefault  = 50000
	CpuPeriodDefault = 100000
)

const (
	cgroupPath       = "/sys/fs/cgroup/mycontainer"
	cgroupSubControl = "/sys/fs/cgroup/mycontainer/cgroup.subtree_control"
	controlGroups    = "+cpu +memory +pids"
)

func NewCgroupContainer() *CgroupContainer {
}

func writeCgroups(config *CgroupConfig, path string) error {
	memory := "max"
	if config.MemoryLimit > 0 {
		memBytes := uint64(config.MemoryLimit * 1024 * 1024)
		memory = strconv.FormatUint(memBytes, 10)
	}

	err := os.WriteFile(filepath.Join(path, "memory.max"), []byte(memory), 0o644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(path, "pids.max"), []byte(strconv.FormatInt(config.PidLimit, 10)), 0o644)
	if err != nil {
		return err
	}

	cpumax := fmt.Sprintf("%d %d", config.CpuQuota, config.CpuPeriod)
	err = os.WriteFile(filepath.Join(path, "cpu.max"), []byte(cpumax), 0o644)
	return err
}

func writePidToCgroups(pid int, path string) error {
	strPid := strconv.Itoa(pid)
	err := os.WriteFile(path, []byte(strPid), 0o644)
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
	var cpuQuota, cpuPeriod int64
	cpuPeriod = CpuPeriodDefault

	switch {
	case cpu > CpuMaxPercentage:
		cpuQuota = CpuQuotaDefault
	case cpu < CpuMinPercentage:
		cpuQuota = CpuQuotaDefault
	}

	if cpuQuota == 0 {
		cpuQuota = (cpu * CpuPeriodDefault) / 100
	}

	return cpuQuota, cpuPeriod
}

func writeCgroupControl() error {
	control := []byte(controlGroups)
	err := os.WriteFile(cgroupSubControl, control, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func cgroupControlExist() (bool, error) {
	data, err := os.ReadFile(cgroupSubControl)
	if err != nil {
		return false, err
	}
	control := string(data)

	subControls := strings.Fields(control)

	expected := []string{"memory", "pids", "cpu"}

	for _, ctr := range expected {
		if !slices.Contains(subControls, ctr) {
			return false, nil
		}
	}

	return true, nil
}

func readCgroupPids(cgroupPath string) ([]int, error) {
	cgroupProc := filepath.Join(cgroupPath, "cgroup.procs")
	data, err := os.ReadFile(cgroupProc)
	if err != nil {
		return nil, err
	}

	str := string(data)

	var pids []int
	for _, line := range strings.Fields(str) {
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	return pids, nil
}

func signalCgroups(cgroupPath string, signal syscall.Signal) error {
	pids, err := readCgroupPids(cgroupPath)
	if err != nil {
		return err
	}

	for _, pid := range pids {
		err = syscall.Kill(pid, signal)
		if err != nil && err != syscall.ESRCH {
			return err
		}
	}

	return nil
}

func cgroupEmpty(cgroupPath string) (bool, error) {
	pids, err := readCgroupPids(cgroupPath)
	if err != nil {
		return false, err
	}

	return len(pids) == 0, nil
}

func killCgroup(cgroupPath string) error {
	cgroupKill := filepath.Join(cgroupPath, "cgroup.kill")
	err := os.WriteFile(cgroupKill, []byte("1"), 0o200)
	if err != nil {
		return err
	}

	return nil
}

func terminateProcess(cgroupPath string, timeout time.Duration) error {
	err := signalCgroups(cgroupPath, syscall.SIGTERM)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		isEmpty, err := cgroupEmpty(cgroupPath)
		if err != nil {
			return err
		}

		if isEmpty {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	err = killCgroup(cgroupPath)
	return err
}
