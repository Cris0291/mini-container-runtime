package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	run := pflag.NewFlagSet("run", pflag.ContinueOnError)

	cmd := run.StringP("cmd", r, "/bin/sh", "command to run")
	rootfs := run.StringP("rootfs", "f", "/rootfs", "path to root")

	run.Parse()
	viper.BindPFlags(run)
}
