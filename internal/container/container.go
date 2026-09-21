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
	fifo         = "exec.fifo"
	lock         = "lock"
	state        = "state.json"
)

type Container struct {
	ContainerDirPath   string
	ContainerPath      string
	ContainerID        string
	BundlePath         string
	ContainerStatePath string
	ContainerLockPath  string
	ContainerFifoPath  string
	ContainerConfig    config.ContainerConfig
}

func NewContainer(containerID string, bundle string) *Container {
	containerPath := filepath.Join(containerDir, containerID)
	containerStatePath := filepath.Join(containerPath, state)
	containerLockPath := filepath.Join(containerPath, lock)
	containerFifoPath := filepath.Join(containerPath, fifo)

	c := &Container{
		ContainerDirPath: containerDir, ContainerID: containerID, ContainerPath: containerPath,
		ContainerStatePath: containerStatePath, BundlePath: bundle,
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

func (container *Container) canonicalizePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	resPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return "", err
	}

	return resPath, nil
}

func Unmarshal[T any](data []byte) (T, error) {
	var state T

	err := json.Unmarshal(data, &state)
	if err != nil {
		return state, err
	}

	return state, nil
}

func validateBundle(bundlePath string) error {
	rootfsPath := filepath.Join(bundlePath, "rootfs")
	info, err := os.Stat(rootfsPath)
	if err != nil {
		return err
	}
}
