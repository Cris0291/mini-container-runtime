package main

import (
	"encoding/json"
	"errors"
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

	childPID := strconv.Itoa(state.PID)
	// i need to check the atomicity of this since the state might change between check
	if state.Status == "stopped" {
		_, err = os.Stat("/proc/" + childPID)
		if err == nil {
			return errors.New("process is currently running stop it first")
		}
		err = os.RemoveAll(containerPath)
		if err != nil {
			return err
		}
	} else {
		_, err = os.Stat("/proc/" + childPID)
		if err != nil {
			err = os.RemoveAll(containerPath)
			if err != nil {
				return err
			}
		} else {
			return errors.New("process is currently running stop it first")
		}
	}

	return nil
}
