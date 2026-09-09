package main

import "path/filepath"

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
