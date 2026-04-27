package main

import (
	"fmt"
	"os"
)

func main() {
	osArgs := os.Args[1:]
	lifeCycleCommand := osArgs[0]
	jsonConfig := osArgs[2]

	switch lifeCycleCommand {
	case "create":
		err := create(jsonConfig)
		fmt.Errorf("create error: ", err)
		os.Exit(1)
	case "run":
		run()
	case "start":
		start()
	case "init":
		runInit()
	}
}
