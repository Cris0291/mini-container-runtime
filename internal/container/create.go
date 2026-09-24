package container

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"containerruntime/internal/cgroup"
	"containerruntime/internal/config"
)

func (container *Container) create(cgroup cgroup.CgroupContainer) (*exec.Cmd, error) {
	configJsonPath := filepath.Join(container.BundlePath, "config.json")

	jsonConfig, err := os.ReadFile(configJsonPath)
	if err != nil {
		return nil, err
	}

	containerConfig, err := Unmarshal[config.ContainerConfig](jsonConfig)
	if err != nil {
		return nil, err
	}

	err = container.validate()
	if err != nil {
		return nil, err
	}

	isControl, err := cgroup.CgroupControlExist()
	if err != nil {
		return nil, err
	}

	// assume that if path exist already cgroup was already written
	if !isControl {
		err = cgroup.WriteCgroupControl()
		if err != nil {
			return nil, err
		}
	}

	cgroupDir := filepath.Join(cgroupPath, config.ID)
	isPath, err = pathExist(cgroupDir)
	if err != nil {
		return nil, err
	}

	if !isPath {
		err = createDir(cgroupDir, 0o700)
		if err != nil {
			return nil, err
		}
	}

	cgroupConfig := normalizeCgroup(config.Resources)
	err = writeCgroups(&cgroupConfig, cgroupDir)
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

	err = writePidToCgroups(cmd.Process.Pid, filepath.Join(cgroupDir, "cgroup.procs"))
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
