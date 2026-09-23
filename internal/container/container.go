package container

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"syscall"

	"containerruntime/internal/config"
)

const (
	containerDir = "/run/mycontainer"
	fifo         = "exec.fifo"
	lock         = "lock"
	state        = "state.json"
)

var validMapSource = []string{"proc", "tmpfs", "sysfs", "devpts", "mqueue"}

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

func NewContainer(containerID string, bundle string) (*Container, error) {
	containerPath := filepath.Join(containerDir, containerID)
	containerStatePath := filepath.Join(containerPath, state)
	containerLockPath := filepath.Join(containerPath, lock)
	containerFifoPath := filepath.Join(containerPath, fifo)

	if !validateID(containerID) {
		return nil, errors.New("invalid container id")
	}

	err := validateBundle(bundle)
	if err != nil {
		return nil, err
	}

	c := &Container{
		ContainerDirPath: containerDir, ContainerID: containerID, ContainerPath: containerPath,
		ContainerStatePath: containerStatePath, BundlePath: bundle,
		ContainerFifoPath: containerFifoPath, ContainerLockPath: containerLockPath,
	}

	return c, nil
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

	if !info.IsDir() {
		return errors.New("rootfs is not a directory")
	}

	entries, err := os.ReadDir(rootfsPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}

		if entryInfo.Mode().Perm()&0o111 != 0 {
			return nil
		}
	}

	return errors.New("no executable was found in rootfs")
}

func validateID(contianerID string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(contianerID)
}

func (container *Container) MountVirtualFileSystems() error {
	rootfsPath := filepath.Join(container.BundlePath, "rootfs")

	for _, mount := range container.ContainerConfig.Mounts {
		path := filepath.Join(rootfsPath, mount.Destination)
		newPath, err := container.canonicalizePath(path)
		if err != nil {
			return err
		}

		// for now bind type is nto allowed
		if !slices.Contains(validMapSource, mount.Type) {
			return errors.New("invalid mount type")
		}

		err = os.MkdirAll(path, 0o711)
		if err != nil {
			return err
		}
		err = syscall.Mount(mount.Source, newPath, mount.Type, uintptr(mount.Flags), mount.Data)
		if err != nil {
			return err
		}
	}
	return nil
}
