package main

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"syscall"
)

func child_init() error {
	envFileDescriptor := os.Getenv("_MYCONTAINER_CONFIGPIPE")
	fd, err := strconv.Atoi(envFileDescriptor)
	if err != nil {
		return err
	}

	file := os.NewFile(uintptr(fd), "config-pipe")
	defer file.Close()

	content, err := io.ReadAll(file)

	var config ContainerConfig

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil
	}

	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
}
