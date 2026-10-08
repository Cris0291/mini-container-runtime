package container

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"containerruntime/internal/config"
)

func (container *Container) ChildInit() error {
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

	containerConfig, err := Unmarshal[config.ContainerConfig](content)
	if err != nil {
		return err
	}

	container.ContainerConfig = containerConfig

	err = syscall.Sethostname([]byte(containerConfig.Hostname))
	if err != nil {
		return fmt.Errorf("set host name step: %w", err)
	}

	// Mount all virtuall filesystems
	err = container.MountVirtualFileSystems()
	if err != nil {
		return fmt.Errorf("mount virtual file system step: %w", err)
	}

	fileExecFifo, err := os.OpenFile(execFifoPath, syscall.O_WRONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("exec fifo step: %w", err)
	}

	err = container.PivotRoot()
	if err != nil {
		return fmt.Errorf("pivot root: %w", err)
	}

	err = config.MountDev()
	if err != nil {
		return err
	}

	err = config.CreateDevNodes()
	if err != nil {
		return err
	}

	err = os.Chdir(container.ContainerConfig.Process.Cwd)
	if err != nil {
		return fmt.Errorf("chdir step: %w", err)
	}

	_, err = fileExecFifo.Write([]byte("0"))
	if err != nil {
		return fmt.Errorf("file exec fifo step: %w", err)
	}

	file.Close()

	path, err := exec.LookPath(container.ContainerConfig.Process.Args[0])
	if err != nil {
		path = container.ContainerConfig.Process.Args[0]
	}

	err = syscall.Exec(path, container.ContainerConfig.Process.Args, container.ContainerConfig.Process.Env)
	if err != nil {
		panic(fmt.Sprintf("exec failed: %v", err))
	}

	return nil
}
