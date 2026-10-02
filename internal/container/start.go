package container

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"syscall"

	"containerruntime/internal/config"
)

func (container *Container) Start() error {
	fileLock, err := container.SetFlock(0, syscall.LOCK_EX)
	if err != nil {
		return err
	}

	defer fileLock.Close()

	data, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return err
	}

	state, err := Unmarshal[config.ContainerState](data)
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

	file, err := os.OpenFile(container.ContainerFifoPath, os.O_RDONLY, 0)
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

	err = os.WriteFile(container.ContainerStatePath, stateData, 0o644)
	if err != nil {
		return err
	}

	return nil
}
