package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func writeStopState(state *ContainerState, statePath *string) error {
	state.Status = "stopped"
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	err = os.WriteFile(*statePath, data, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func readCgroupPids(cgroupPath *string) ([]int, error) {
	data, err := os.ReadFile(*cgroupPath)
	if err != nil {
		return nil, err
	}

	str := string(data)

	var pids []int
	for _, line := range strings.Fields(str) {
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	return pids, nil
}

func signalCgroups(cgroupPath *string, signal syscall.Signal) error {
	pids, err := readCgroupPids(cgroupPath)
	if err != nil {
		return err
	}

	for _, pid := range pids {
		err = syscall.Kill(pid, signal)
		if err != nil && err != syscall.ESRCH {
			return err
		}
	}

	return nil
}

func cgroupEmpty(cgroupPath *string) (bool, error) {
	pids, err := readCgroupPids(cgroupPath)
	if err != nil {
		return false, err
	}

	return len(pids) == 0, nil
}

func killCgroup(cgroupPath *string) error {
	err := os.WriteFile(*cgroupPath, []byte("1"), 0o200)
	if err != nil {
		return err
	}

	return nil
}

func terminateProcess(cgroupPath *string, timeout time.Duration) error {
	err := signalCgroups(cgroupPath, syscall.SIGTERM)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		isEmpty, err := cgroupEmpty(cgroupPath)
		if err != nil {
			return err
		}

		if isEmpty {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	err = killCgroup(cgroupPath)
	return err
}

func stop(containerID string) error {
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")
	lockPath := filepath.Join(containerPath, "lock")
	cgroupPath := "/sys/fs/cgroup/mycontainer"
	cgroupDir := filepath.Join(cgroupPath, containerID)

	fileLock, err := os.OpenFile(lockPath, syscall.O_RDWR, 0)
	if err != nil {
		return err
	}

	err = syscall.Flock(int(fileLock.Fd()), syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer fileLock.Close()

	file, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}

	var state ContainerState
	err = json.Unmarshal(file, &state)
	if err != nil {
		return err
	}

	if state.Status == "stopped" {
		return nil
	}

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err == nil {
		err = terminateProcess(&cgroupDir, 10*time.Second)
		if err != nil {
			// if we reach this path means that p.kill went worng which should not happen
			panic("process could not be killed os has failed us")
		}
	}
	err = writeStopState(&state, &statePath)
	// here is the problem if i return the error here it means that the process is dead
	// but i could not write the state so the semantics are weird operation was successful but the result is missleading
	if err != nil {
		return fmt.Errorf("process stopped but failed to persist the state: %w", err)
	}
	return nil
}
