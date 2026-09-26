package container

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
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

var NamespaceRelation = map[string]uintptr{
	"pid":    syscall.CLONE_NEWPID,
	"uts":    syscall.CLONE_NEWUTS,
	"mount":  syscall.CLONE_NEWNS,
	"net":    syscall.CLONE_NEWNET,
	"ipc":    syscall.CLONE_NEWIPC,
	"user":   syscall.CLONE_NEWUSER,
	"cgroup": syscall.CLONE_NEWCGROUP,
}

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

	err = config.MakeDir(containerPath, 0o700)
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

func NewEmptyContainer() *Container {
	c := &Container{}
	return c
}

func (container *Container) SetFlock(perm os.FileMode, flockHow int) (*os.File, error) {
	fileLock, err := os.OpenFile(container.ContainerLockPath, syscall.O_RDWR, perm)
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
	for _, mount := range container.ContainerConfig.Mounts {
		path := filepath.Join(container.ContainerConfig.Rootfs, mount.Destination)
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

func (container *Container) PivotRoot() error {
	syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	err := syscall.Mount(container.ContainerConfig.Rootfs, container.ContainerConfig.Rootfs, "", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	pivotDir := filepath.Join(container.ContainerConfig.Rootfs, ".pivot_root")
	err = os.MkdirAll(pivotDir, 0o711)
	if err != nil {
		return err
	}

	err = syscall.PivotRoot(container.ContainerConfig.Rootfs, pivotDir)
	if err != nil {
		return err
	}

	err = os.Chdir("/")
	if err != nil {
		return err
	}

	err = syscall.Unmount("/.pivot_root", syscall.MNT_DETACH)
	if err != nil {
		return err
	}

	err = os.Remove("/.pivot_root")
	if err != nil {
		return err
	}

	return nil
}

func (container *Container) validate() error {
	if container.ContainerConfig.Hostname == "" {
		return errors.New("no hostname was provided i the json config file")
	}
	return nil
}

func (container *Container) SetState(state config.ContainerConfig) {
	container.ContainerConfig = state
	container.ContainerConfig.Rootfs = filepath.Join(container.BundlePath, "rootfs")
}

func (container *Container) CloneFlags() uintptr {
	var flags uintptr
	for _, namespace := range container.ContainerConfig.Namespaces {
		if strings.TrimSpace(namespace.Path) == "" {
			value, ok := NamespaceRelation[namespace.Type]
			if ok {
				flags |= value
			}
		}
	}
	return flags
}
