package config

import (
	"errors"
	"fmt"
	"os"
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

var (
	MYCONTAINER_CONFIGPIPE = "_MYCONTAINER_CONFIGPIPE=3"
	MYCONTAINER_EXECFIFO   = "_MYCONTAINER_EXECFIFO=4"
	MYCONTAINER_CONFIGID   = "_MYCONTAINER_CONFIGID="
)

func CreateDir(path string, perm os.FileMode) error {
	err := os.Mkdir(path, perm)
	if err != nil {
		return err
	}
	return nil
}

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func MakeDir(path string, perm os.FileMode) error {
	isPath, err := PathExist(path)
	if err != nil {
		return err
	}
	if !isPath {
		err = CreateDir(path, perm)
		if err != nil {
			return err
		}
	}

	return nil
}

func MountDev() error {
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

func CreateDevNodes() error {
	for path, majmin := range globalDeviceMap {
		err := syscall.Mknod(path, syscall.S_IFCHR|0o666, makedev(majmin[0], majmin[1]))
		if err != nil {
			return fmt.Errorf("in create dev nodes: %w", err)
		}
	}
	return nil
}
