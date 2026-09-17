package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var namespaceRelation = map[string]uintptr{
	"pid":    syscall.CLONE_NEWPID,
	"uts":    syscall.CLONE_NEWUTS,
	"mount":  syscall.CLONE_NEWNS,
	"net":    syscall.CLONE_NEWNET,
	"ipc":    syscall.CLONE_NEWIPC,
	"user":   syscall.CLONE_NEWUSER,
	"cgroup": syscall.CLONE_NEWCGROUP,
}

var globalDeviceMap = map[string][2]uint32{
	"/dev/null":    {1, 3},
	"/dev/zero":    {1, 5},
	"/dev/full":    {1, 7},
	"/dev/random":  {1, 8},
	"/dev/urandom": {1, 9},
	"/dev/tty":     {5, 0},
}

func (c *ContainerConfig) CloneFlags() uintptr {
	var flags uintptr
	for _, namespace := range c.Namespaces {
		if strings.TrimSpace(namespace.Path) == "" {
			value, ok := namespaceRelation[namespace.Type]
			if ok {
				flags |= value
			}
		}
	}
	return flags
}

func (config *ContainerConfig) validate() error {
	if config.ID == "" {
		return errors.New("no id was provided in the json file")
	}
	if config.Hostname == "" {
		return errors.New("no hostname was provided i the json config file")
	}
	if config.Rootfs == "" {
		return errors.New("no rootfs was provided in the json config file")
	}
	_, err := os.Stat(config.Rootfs)
	if err != nil {
		return errors.New("rootfs path does not exist")
	}
	return nil
}

func createDir(path string, perm os.FileMode) error {
	err := os.Mkdir(path, perm)
	if err != nil {
		return err
	}
	return nil
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func (config *ContainerConfig) MountVirtualFileSystems() error {
	for _, mount := range config.Mounts {
		path := filepath.Join(config.Rootfs, mount.Destination)
		err := os.MkdirAll(path, 0o711)
		if err != nil {
			return err
		}
		err = syscall.Mount(mount.Source, path, mount.Type, uintptr(mount.Flags), mount.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (config *ContainerConfig) PivotRoot() error {
	syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	err := syscall.Mount(config.Rootfs, config.Rootfs, "", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	pivotDir := filepath.Join(config.Rootfs, ".pivot_root")
	err = os.MkdirAll(pivotDir, 0o711)
	if err != nil {
		return err
	}

	err = syscall.PivotRoot(config.Rootfs, pivotDir)
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

func mountDev() error {
	_, err := os.Stat("/dev")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		os.Mkdir("/dev", 0o755)
	} else if err != nil {
		return err
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID, "mode=755")
	if err != nil {
		return err
	}

	return nil
}

func makedev(major, minor uint32) int {
	return int((major << 8) | minor)
}

func createDevNodes() error {
	for path, majmin := range globalDeviceMap {
		err := syscall.Mknod(path, syscall.S_IFCHR|0o666, makedev(majmin[0], majmin[1]))
		if err != nil {
			return fmt.Errorf("in create dev nodes: %w", err)
		}
	}
	return nil
}
