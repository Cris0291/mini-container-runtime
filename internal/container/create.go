package container

import "os/exec"

func (contianer *Container) create(pathConfig string) (*exec.Cmd, error) {
	// this path should not be in the json config
	// it should be dynamically created the mycontainer part is temporary
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

	isPath, err := pathExist(cgroupPath)
	if err != nil {
		return nil, err
	}

	if !isPath {
		err = createDir(cgroupPath, 0o700)
		if err != nil {
			return nil, err
		}
	}

	isControl, err := cgroupControlExist()
	if err != nil {
		return nil, err
	}

	// assume that if path exist already cgroup was already written
	if !isControl {
		err = writeCgroupControl()
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
