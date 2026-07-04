package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func start(containerID string) error {
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")
	lockPath := filepath.Join(containerPath, "lock")

	fileLock, err := os.OpenFile(lockPath, syscall.O_RDWR, 0)
	if err != nil {
		return err
	}

	err = syscall.Flock(int(fileLock.Fd()), syscall.LOCK_EX)
	if err != nil {
		return err
	}

	defer fileLock.Close()

	data, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return err
	}

	if state.Status != "created" {
		return errors.New("something unexpeted happened status different from created")
	}

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err != nil {
		return err
	}

	execFifoPath := filepath.Join(containerPath, "exec.fifo")
	file, err := os.OpenFile(execFifoPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	defer file.Close()

	if err != nil {
		return err
	}

	buff := make([]byte, 1)
	_, err = file.Read(buff)
	if err != nil {
		return err
	}

	state.Status = "running"
	stateData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = os.WriteFile(statePath, stateData, 0o644)
	if err != nil {
		return err
	}

	return nil
}
