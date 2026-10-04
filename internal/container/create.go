package container

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"containerruntime/internal/cgroup"
	"containerruntime/internal/config"
)

func (container *Container) Create(cgroup *cgroup.CgroupContainer) (*exec.Cmd, error) {
	// this path should not be in the json config
	// it should be dynamically created the mycontainer part is temporary
	path := filepath.Join(container.BundlePath, "config.json")

	jsonConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	containerConfig, err := Unmarshal[config.ContainerConfig](jsonConfig)
	if err != nil {
		return nil, err
	}

	container.SetState(containerConfig)

	err = container.validate()
	if err != nil {
		return nil, err
	}

	isControl, err := cgroup.CgroupControlExist()
	if err != nil {
		return nil, err
	}

	if !isControl {
		err = cgroup.WriteCgroupControl()
		if err != nil {
			return nil, err
		}
	}

	err = config.MakeDir(cgroup.ContainerPath, 0o700)
	if err != nil {
		return nil, err
	}

	cgroup.NormalizeCgroup(containerConfig.Resources)
	err = cgroup.WriteCgroups()
	if err != nil {
		return nil, err
	}

	fileLock, err := os.OpenFile(container.ContainerLockPath, syscall.O_RDWR|syscall.O_CREAT, 0o666)
	if err != nil {
		return nil, err
	}

	defer fileLock.Close()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = append(cmd.ExtraFiles, r)
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: container.CloneFlags(), Setsid: true}

	cmd.Env = append(cmd.Env, config.MYCONTAINER_CONFIGPIPE)

	err = syscall.Mkfifo(container.ContainerFifoPath, 0o622)
	if err != nil {
		return nil, err
	}

	cmd.Env = append(cmd.Env, config.MYCONTAINER_CONFIGID+container.ContainerID)

	cmd.Env = append(cmd.Env, config.MYCONTAINER_EXECFIFO)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cgroup.WritePidToCgroups(cmd.Process.Pid)
	if err != nil {
		return nil, err
	}

	r.Close()

	state := config.ContainerState{
		ID:      container.ContainerID,
		PID:     cmd.Process.Pid,
		Status:  "created",
		Bundle:  container.BundlePath,
		Created: time.Now().UTC(),
		Config:  container.ContainerConfig,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(container.ContainerConfig)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(configData)
	if err != nil {
		return nil, err
	}

	w.Close()

	err = os.WriteFile(container.ContainerStatePath, data, 0o644)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
