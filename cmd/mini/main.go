package main

import (
	"os"
)

func main() {
	lifeCycleCommand := os.Args[1:][0]
	switch lifeCycleCommand {
	case "create":
		create()
	case "run":
		run()
	case "start":
		start()
	case "init":
		runInit()
	}
}
