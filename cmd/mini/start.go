package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func start() error {
	containerID := os.Args[0]
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	json.Unmarshal(data, &state)

	execFifoPath := filepath.Join(containerPath, "exec.fifo")

	return nil
}
