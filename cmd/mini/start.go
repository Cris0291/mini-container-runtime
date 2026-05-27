package main

import (
	"encoding/json"
	"errors"
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

	if state.Status != "created" {
		return errors.New("something unexpeted happened status different from created")
	}

	execFifoPath := filepath.Join(containerPath, "exec.fifo")
	file, err := os.OpenFile(execFifoPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	defer file.Close()

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
