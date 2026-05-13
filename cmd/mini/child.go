package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func MountVirtualFileSystems(config *ContainerConfig) error {
	for _, mount := range config.Mounts {
		path := filepath.Join(config.Rootfs, mount.Destination)
		err := os.Mkdir(path, 0o711)
		if err != nil {
			return err
		}
		err = syscall.Mount(mount.Source, path, mount.Source, uintptr(mount.Flags), mount.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func PivotRoot(config *ContainerConfig) error {
	syscall.Mount(config.Rootfs, config.Rootfs, "", syscall.MS_BIND|syscall.MS_REC, "")

	pivotDir := filepath.Join(config.Rootfs, ".pivot_root")
	err := os.Mkdir(pivotDir, 0o711)
	if err != nil {
		return err
	}

	err = syscall.PivotRoot(config.Rootfs, "/")
	if err != nil {
		return err
	}

	os.Chdir("/")

	syscall.Unmount("/.pivot_root", syscall.MNT_DETACH)

	os.Remove("/.pivot_root")

	return nil
}

func ChildInit() error {
	envFileDescriptor := os.Getenv("_MYCONTAINER_CONFIGPIPE")
	fd, err := strconv.Atoi(envFileDescriptor)
	if err != nil {
		return err
	}

	file := os.NewFile(uintptr(fd), "config-pipe")
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var config ContainerConfig

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil
	}

	err = syscall.Sethostname([]byte(config.Hostname))
	if err != nil {
		return err
	}

	// Mount all virtuall filesystems
	err = MountVirtualFileSystems(&config)
	if err != nil {
		return err
	}

	return nil
}
