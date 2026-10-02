package main

import (
	"fmt"
	"os"

	"containerruntime/internal/cgroup"
	"containerruntime/internal/container"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "too few arguments")
		os.Exit(1)
	}

	var containerObject *container.Container
	var cgroupObject *cgroup.CgroupContainer
	var containerID string
	var bundlePath string

	lifeCycleCommand := os.Args[1]
	if lifeCycleCommand != "child" {
		containerID = os.Args[2]
		bundlePath = os.Args[3]
		var err error

		containerObject, err = container.NewContainer(containerID, bundlePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "container init erro %v\n", err)
		}
		cgroupObject, err = cgroup.NewCgroupContainer(containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cgroup container init erro %v\n", err)
		}
	} else {
		containerObject = container.NewEmptyContainer()
	}

	switch lifeCycleCommand {
	case "create":
		_, err := containerObject.Create(cgroupObject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create error %v\n", err)
			return
		}
	case "run":
		err := containerObject.Run(cgroupObject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run error %v\n", err)
		}
	case "start":
		err := containerObject.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "start error %v\n", err)
			return
		}
	case "child":
		err := containerObject.ChildInit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "child error %v\n", err)
			return
		}
	case "delete":
		err := containerObject.Delete(cgroupObject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete error %v\n", err)
		}
	case "stop":
		err := containerObject.Stop(cgroupObject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stop error %v\n", err)
		}
	case "state":
		err := containerObject.State()
		if err != nil {
			fmt.Fprintf(os.Stderr, "state error %v\n", err)
		}
	}
}
