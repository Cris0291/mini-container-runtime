package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
)

type Container struct {
	ContainerDirPath   string
	ContainerPath      string
	ContainerID        string
	ContainerStatePath string
	ContainerLockPath  string
	ContainerFifoPath  string
	statePath          string
	lockPath           string
	fifoPath           string
}

func NewContainer(containerDir string, containerID string, state string, lock string, fifo string) *Container {
	c := &Container{ContainerDirPath: containerDir, ContainerID: containerID, statePath: state, lockPath: lock, fifoPath: fifo}
	c.initialize()
	return c
}

func (container *Container) initialize() {
	container.ContainerPath = filepath.Join(container.ContainerDirPath, container.ContainerID)
	container.ContainerStatePath = filepath.Join(container.ContainerPath, container.statePath)
	container.ContainerLockPath = filepath.Join(container.ContainerPath, container.lockPath)
	container.ContainerFifoPath = filepath.Join(container.ContainerPath, container.fifoPath)
}

func (container *Container) SetFlock(flockHow int) (*os.File, error) {
	fileLock, err := os.OpenFile(container.ContainerLockPath, syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(fileLock.Fd()), flockHow)
	if err != nil {
		return nil, err
	}

	return fileLock, nil
}

func (container *Container) CreateContainerState() (*ContainerState, error) {
	file, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return nil, err
	}
	var state ContainerState
	err = json.Unmarshal(file, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
