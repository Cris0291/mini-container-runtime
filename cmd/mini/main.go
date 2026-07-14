package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "too few arguments")
		os.Exit(1)
	}

	lifeCycleCommand := os.Args[1]

	switch lifeCycleCommand {
	case "create":
		jsonConfig := os.Args[2]
		_, err := create(jsonConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create error %v\n", err)
			return
		}
	case "run":
		jsonConfig := os.Args[2]
		containerID := os.Args[3]
		err := run(jsonConfig, containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run error %v\n", err)
		}
	case "start":
		containerID := os.Args[2]
		err := start(containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "start error %v\n", err)
			return
		}
	case "init":
		runInit()
	case "child":
		err := ChildInit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "child error %v\n", err)
			return
		}
	case "delete":
		containerID := os.Args[2]
		err := delete(containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete error %v\n", err)
		}
	case "stop":
		containerID := os.Args[2]
		err := stop(containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stop error %v\n", err)
		}
	case "state":
		containerID := os.Args[2]
		err := state(containerID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "state error %v\n", err)
		}
	}
}
