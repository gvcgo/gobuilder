package main

import (
	"os"

	"github.com/gvcgo/gobuild/cmd"
	"github.com/gvcgo/gobuild/internal"
)

func main() {
	cwd, _ := os.Getwd()
	if cwd != "" {
		internal.SetCurrentWorkingDir(cwd)
	}
	cli := cmd.NewCli()
	cli.Run()
}
