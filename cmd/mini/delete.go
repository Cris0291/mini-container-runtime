package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

func delete(containerID string) error {
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")

	data, err := os.ReadFile(statePath)

	var state ContainerState
	err = json.Unmarshal(data, state)
	if err != nil {
		return err
	}

	// i need to check the atomicity of this since the state might change between check
	if state.Status == "running" {
		childPID := strconv.Itoa(state.PID)
		os.Stat("/proc/" + childPID)
	}
}
