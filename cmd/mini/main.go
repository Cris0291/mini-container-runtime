package main

import (
	"os"
)

func main() {
	osArgs := os.Args[1:]
	lifeCycleCommand := osArgs[0]
	jsonConfig := os.Args[1]

	switch lifeCycleCommand {
	case "create":
		create(jsonConfig)
	case "run":
		run()
	case "start":
		start()
	case "init":
		runInit()
	}
}
