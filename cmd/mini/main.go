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
		err := create(jsonConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create error %v\n", err)
			return
		}
	case "run":
		run()
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
		delete()
	}
}
