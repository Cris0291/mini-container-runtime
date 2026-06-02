package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		panic("too few arguments")
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
		start(containerID)
	case "init":
		runInit()
	case "child":
		ChildInit()
	}
}
