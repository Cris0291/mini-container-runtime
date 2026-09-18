package container

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"

	"containerruntime/internal/config"
)

type Container struct {
	ContainerDirPath        string
	ContainerPath           string
	ContainerID             string
	ContainerConfigJsonPath string
	ContainerStatePath      string
	ContainerLockPath       string
	ContainerFifoPath       string
}

func NewContainer(containerDir string, containerID string, jsonPath string, state string, lock string, fifo string) *Container {
	containerPath := filepath.Join(containerDir, containerID)
	containerStatePath := filepath.Join(containerPath, state)
	containerLockPath := filepath.Join(containerPath, lock)
	containerFifoPath := filepath.Join(containerPath, fifo)
	containerConfigJsonPath := filepath.Join(containerPath, jsonPath)

	c := &Container{
		ContainerDirPath: containerDir, ContainerID: containerID, ContainerPath: containerPath,
		ContainerStatePath: containerStatePath, ContainerConfigJsonPath: containerConfigJsonPath,
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
