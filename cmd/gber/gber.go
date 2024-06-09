package main

import (
	"github.com/gvcgo/gobuilder/cmd"
)

var (
	GitTag  string
	GitHash string
)

func main() {
	cli := cmd.NewCli(GitTag, GitHash)
	cli.Run()
}
