package container

import (
	"errors"
	"os"
	"strconv"
	"syscall"

	"containerruntime/internal/cgroup"
	"containerruntime/internal/config"
)

func (container *Container) Delete(cgroup *cgroup.CgroupContainer) error {
	fileLock, err := container.SetFlock(0, syscall.LOCK_EX)
	if err != nil {
		return err
	}

	defer fileLock.Close()

	data, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return err
	}

	state, err := Unmarshal[config.ContainerState](data)
	if err != nil {
		return err
	}

	childPID := strconv.Itoa(state.PID)
	// i need to check the atomicity of this since the state might change between check
	if state.Status == "stopped" {
		_, err = os.Stat("/proc/" + childPID)
		if err == nil {
			return errors.New("process is currently running stop it first")
		}
		err = os.RemoveAll(container.ContainerPath)
		if err != nil {
			return err
		}
		err = os.Remove(cgroup.ContainerPath)
		if err != nil {
			return err
		}

	} else {
		_, err = os.Stat("/proc/" + childPID)
		if err != nil {
			err = os.RemoveAll(container.ContainerPath)
			if err != nil {
				return err
			}
			err = os.Remove(cgroup.ContainerPath)
			if err != nil {
				return err
			}
		} else {
			return errors.New("process is currently running stop it first")
		}
	}

	return nil
}
