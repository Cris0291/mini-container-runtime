package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func delete(containerID string) error {
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return err
	}

	// i need to check the atomicity of this since the state might change between check
	if state.Status == "stopped" {
		childPID := strconv.Itoa(state.PID)
		_, err = os.Stat("/proc/" + childPID)
		if err == nil {
			return errors.New("process is currently running stop it first")
		}
		err = os.RemoveAll(containerPath)
		if err != nil {
			return err
		}
	} else if state.Status == "delete" {
		return errors.New("process ")
	}

	return nil
}
