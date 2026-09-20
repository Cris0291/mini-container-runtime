package container

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"

	"containerruntime/internal/config"
)

const (
	containerDir = "/run/mycontainer"
)

type Container struct {
	ContainerDirPath   string
	ContainerPath      string
	ContainerID        string
	BundlePath         string
	ContainerStatePath string
	ContainerLockPath  string
	ContainerFifoPath  string
}

func NewContainer(containerID string, bundle string, state string, lock string, fifo string) *Container {
	containerPath := filepath.Join(containerDir, containerID)
	containerStatePath := filepath.Join(containerPath, state)
	containerLockPath := filepath.Join(containerPath, lock)
	containerFifoPath := filepath.Join(containerPath, fifo)

	c := &Container{
		ContainerDirPath: containerDir, ContainerID: containerID, ContainerPath: containerPath,
		ContainerStatePath: containerStatePath, BundlePath: bundle,,
		ContainerFifoPath: containerFifoPath, ContainerLockPath: containerLockPath,
	}

	return c
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

func (container *Container) CreateContainerState() (*config.ContainerState, error) {
	file, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return nil, err
	}
	var state config.ContainerState
	err = json.Unmarshal(file, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func(container *Container) canonicalizePath(path string){
	cleanPath := filepath.Clean(path)
}
