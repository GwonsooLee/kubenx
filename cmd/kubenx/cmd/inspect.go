package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get pod
func NewCmdInspect() *cobra.Command {
	return NewCmd("inspect").
		WithDescription("Inspect resource in detail").
		AddSearchGroups().
		SetAliases([]string{"ins"}).
		AddInspectGroups().
		RunWithArgsAndCmd(execInsepct)
}

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Short:   "A brief description of your command",
	Long:    `Inspect resource in detail`,
	Run:     func(cmd *cobra.Command, args []string) {},
	Aliases: []string{"ins"},
}

// Function for search execution
func execInsepct(_ context.Context, _ io.Writer, cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
