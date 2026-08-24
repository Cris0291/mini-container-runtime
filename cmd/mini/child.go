package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

var globalDeviceMap = map[string][2]uint32{
	"/dev/null":    {1, 3},
	"/dev/zero":    {1, 5},
	"/dev/full":    {1, 7},
	"/dev/random":  {1, 8},
	"/dev/urandom": {1, 9},
	"/dev/tty":     {5, 0},
}

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
	} else {
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
			return err
		}
	}
	return nil
}

func ChildInit() error {
	fmt.Fprintf(os.Stderr, "child started")
	containerID := os.Getenv("_MYCONTAINER_CONFIGID")
	containerPath := filepath.Join("/run/mycontainer", containerID)
	execFifoPath := filepath.Join(containerPath, "exec.fifo")
	envFileDescriptor := os.Getenv("_MYCONTAINER_CONFIGPIPE")
	fd, err := strconv.Atoi(envFileDescriptor)
	if err != nil {
		return fmt.Errorf("atoi step: %w", err)
	}

	file := os.NewFile(uintptr(fd), "config-pipe")
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("config pipe file step: %w", err)
	}

	var config ContainerConfig

	err = json.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("unmarshall step: %w", err)
	}

	err = syscall.Sethostname([]byte(config.Hostname))
	if err != nil {
		return fmt.Errorf("set host name step: %w", err)
	}

	// Mount all virtuall filesystems
	err = MountVirtualFileSystems(&config)
	if err != nil {
		return fmt.Errorf("mount virtual file system step: %w", err)
	}

	err = PivotRoot(&config)
	if err != nil {
		return fmt.Errorf("pivot root: %w", err)
	}

	err = mountDev()
	if err != nil {
		return err
	}

	err = createDevNodes()
	if err != nil {
		return err
	}

	err = os.Chdir(config.Process.Cwd)
	if err != nil {
		return fmt.Errorf("chdir step: %w", err)
	}

	fileExecFifo, err := os.OpenFile(execFifoPath, syscall.O_WRONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("exec fifo step: %w", err)
	}

	_, err = fileExecFifo.Write([]byte("0"))
	if err != nil {
		return fmt.Errorf("file exec fifo step: %w", err)
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
