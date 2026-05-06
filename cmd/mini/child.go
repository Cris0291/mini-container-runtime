package main

import (
	"os"
)

func child_init() {
	fd := os.Getenv("_MYCONTAINER_CONFIGPIPE")
}
