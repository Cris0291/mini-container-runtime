package main

import (
	"fmt"
	"os"
)

func main() {
	osArgs := os.Args[1:]
	lifeCycleCommand := osArgs[0]
	jsonConfig := osArgs[1]

	switch lifeCycleCommand {
	case "create":
		err := create(jsonConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create error %v\n", err)
			return
		}
	case "run":
		run()
	case "start":
		start()
	case "init":
		runInit()
	}
}
