package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func MountVirtualFileSystems(config *ContainerConfig) error {
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

func PivotRoot(config *ContainerConfig) error {
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
		return err
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

	err = PivotRoot(&config)
	if err != nil {
		return err
	}

	err = os.Chdir(config.Process.Cwd)
	if err != nil {
		return err
	}

	namedPipeFileDescriptor := os.Getenv("_MYCONTAINER_EXECFIFO")
	npfd, err := strconv.Atoi(namedPipeFileDescriptor)
	if err != nil {
		return err
	}

	file = os.NewFile(uintptr(npfd), "exec.fifo")
	_, err = file.Write([]byte("0"))
	if err != nil {
		return err
	}
	file.Close()

	path, err := exec.LookPath(config.Process.Args[0])
	if err != nil {
		path = config.Process.Args[0]
	}

	err = syscall.Exec(path, config.Process.Args, config.Process.Env)
	if err != nil {
		panic("something terribly wrong happened")
	}

	return nil
}
