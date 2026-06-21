package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func writeStopState(state *ContainerState, statePath *string) error {
	state.Status = "stopped"
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	err = os.WriteFile(*statePath, data, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func terminateProcess(pid int, timeout time.Duration, state *ContainerState, statePath *string) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		// Process was killed in between just save the stopped state
		err = writeStopState(state, statePath)
		return err
	}

	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		// Process was already kill
		writeStopState(state, statePath)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
	}
}

func stop(containerID string) error {
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

	file, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	err = json.Unmarshal(file, &state)
	if err != nil {
		return err
	}

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err != nil {
		err = WriteStopState(&state, &statePath)
		return err
	}
}
