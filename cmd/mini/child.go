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
