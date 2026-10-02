package container

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"containerruntime/internal/cgroup"
	"containerruntime/internal/config"
)

func (container *Container) Stop(cgroup *cgroup.CgroupContainer) error {
	fileLock, err := container.SetFlock(0, syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer fileLock.Close()

	file, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return err
	}

	state, err := Unmarshal[config.ContainerState](file)
	if err != nil {
		return err
	}

	if state.Status == "stopped" {
		return nil
	}

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err == nil {
		err = cgroup.TerminateProcess(10 * time.Second)
		if err != nil {
			// if we reach this path means that p.kill went worng which should not happen
			panic("process could not be killed os has failed us")
		}
	}
	err = cgroup.WriteStopState(&state, &container.ContainerStatePath)
	// here is the problem if i return the error here it means that the process is dead
	// but i could not write the state so the semantics are weird operation was successful but the result is missleading
	if err != nil {
		return fmt.Errorf("process stopped but failed to persist the state: %w", err)
	}
	return nil
}
