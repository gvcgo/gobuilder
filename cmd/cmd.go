package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gvcgo/gobuilder/internal"
	"github.com/gvcgo/goutils/pkgs/gtea/confirm"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
	"github.com/spf13/cobra"
)

var GroupID string = "gber"

type Cli struct {
	rootCmd *cobra.Command
	gitTag  string
	gitHash string
}

func NewCli(gitTag, gitHash string) (c *Cli) {
	c = &Cli{
		rootCmd: &cobra.Command{
			Short: "A enhanced go builder.",
			Long:  "gber <subcommand> --flags <args>.",
		},
		gitTag:  gitTag,
		gitHash: gitHash,
	}
	c.rootCmd.AddGroup(&cobra.Group{ID: GroupID, Title: "Command list: "})
	c.initiate()
	return
}

func (c *Cli) initiate() {
	c.rootCmd.AddCommand(&cobra.Command{
		Use:                "build",
		Aliases:            []string{"b"},
		Short:              "Builds a go project.",
		Long:               "Example: gber build --flags <args>.",
		GroupID:            GroupID,
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			bd := internal.NewGoBuilder()
			bd.Build()
		},
	})

	c.rootCmd.AddCommand(&cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Shows version info go gber.",
		GroupID: GroupID,
		Run: func(cmd *cobra.Command, args []string) {
			if len(c.gitHash) > 7 {
				c.gitHash = c.gitHash[:8]
			}
			if c.gitTag != "" && c.gitHash != "" {
				fmt.Println(gprint.CyanStr("%s(%s)", c.gitTag, c.gitHash))
			}
		},
	})

	c.rootCmd.AddCommand(&cobra.Command{
		Use:     "clear",
		Aliases: []string{"c"},
		Short:   "Clears the build directory.",
		GroupID: GroupID,
		Run: func(cmd *cobra.Command, args []string) {
			pd := internal.FindGoProjectDir(internal.GetCurrentWorkingDir())
			buildDir := filepath.Join(pd, "build")
			if ok, _ := gutils.PathIsExist(buildDir); ok {
				cfm := confirm.NewConfirmation(confirm.WithPrompt(fmt.Sprintf("Do you really mean to clear %s?", buildDir)))
				cfm.Run()

				if cfm.Result() {
					os.RemoveAll(buildDir)
				}
			}
		},
	})
}

func (c *Cli) Run() {
	if err := c.rootCmd.Execute(); err != nil {
		gprint.PrintError("%+v", err)
	}
}
