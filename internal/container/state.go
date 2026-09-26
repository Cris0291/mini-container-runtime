package container

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"containerruntime/internal/config"
)

func (container *Container) state() error {
	fileLock, err := container.SetFlock(0, syscall.LOCK_SH)

	defer fileLock.Close()

	file, err := os.ReadFile(container.ContainerStatePath)
	if err != nil {
		return err
	}

	state, err := Unmarshal[config.ContainerState](file)

	childPID := strconv.Itoa(state.PID)
	_, err = os.Stat("/proc/" + childPID)
	if err != nil {
		if state.Status == "running" {
			fmt.Println("there is a difference in status process does not exist yet is status is: running")
		} else {
			fmt.Println("process does not exist and its stauts is : " + state.Status)
		}
		state.Status = "stopped"
	}
	res, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}
