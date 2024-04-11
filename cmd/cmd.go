package cmd

import (
	"github.com/gvcgo/gobuilder/internal"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/spf13/cobra"
)

var GroupID string = "gber"

type Cli struct {
	rootCmd *cobra.Command
}

func NewCli() (c *Cli) {
	c = &Cli{
		rootCmd: &cobra.Command{
			Short: "A enhanced go builder.",
			Long:  "gber <subcommand> --flags <args>.",
		},
	}
	c.rootCmd.AddGroup(&cobra.Group{ID: GroupID, Title: "Command list: "})
	c.initiate()
	return
}

func (c *Cli) initiate() {
	c.rootCmd.AddCommand(&cobra.Command{
		Use:                "build",
		Aliases:            []string{"b"},
		Short:              "Build a go project.",
		Long:               "Example: gber build --flags <args>.",
		GroupID:            GroupID,
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			bd := internal.NewGoBuilder()
			bd.Build()
		},
	})
}

func (c *Cli) Run() {
	if err := c.rootCmd.Execute(); err != nil {
		gprint.PrintError("%+v", err)
	}
}
