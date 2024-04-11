package main

import (
	"os"

	"github.com/gvcgo/gobuilder/cmd"
	"github.com/gvcgo/gobuilder/internal"
)

func main() {
	cwd, _ := os.Getwd()
	if cwd != "" {
		internal.SetCurrentWorkingDir(cwd)
	}

	cli := cmd.NewCli()
	cli.Run()
}
