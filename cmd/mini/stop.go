package main

import (
	"encoding/json"
	"fmt"
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

func terminateProcess(pid int, timeout time.Duration) error {
	p, err := os.FindProcess(pid)

	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		// Process was already kill
		return nil
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err = syscall.Kill(pid, 0)
		if err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return p.Kill()
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
	defer fileLock.Close()

	file, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	err = json.Unmarshal(file, &state)
	if err != nil {
		return err
	}

	if state.Status == "stopped" {
		return nil
	}

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err == nil {
		err = terminateProcess(state.PID, 10*time.Second)
		if err != nil {
			// if we reach this path means that p.kill went worng which should not happen
			panic("process could not be killed os has failed us")
		}
	}
	err = writeStopState(&state, &statePath)
	// here is the problem if i return the error here it means that the process is dead
	// but i could not write the state so the semantics are weird operation was successful but the result is missleading
	if err != nil {
		return fmt.Errorf("process stopped but failed to persist the state: %w", err)
	}
	return nil
}
