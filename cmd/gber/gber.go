package main

import (
	"os"

	"github.com/gvcgo/gobuilder/cmd"
	"github.com/gvcgo/gobuilder/internal"
)

var (
	GitTag  string
	GitHash string
)

func main() {
	cwd, _ := os.Getwd()
	if cwd != "" {
		internal.SetCurrentWorkingDir(cwd)
	}

	cli := cmd.NewCli(GitTag, GitHash)
	cli.Run()
}
