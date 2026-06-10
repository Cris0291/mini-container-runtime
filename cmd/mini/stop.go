package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
)

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
}
