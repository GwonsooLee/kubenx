package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"io"
)

//Get new get function
func NewCmdGet() *cobra.Command {
	return NewCmd("get").
		WithDescription("Get kubernetes information").
		WithLongDescription("Get command for retrieve inforamtion").
		SetAliases([]string{"ge"}).
		AddGetGroups().
		SetFlags().
		RunWithArgsAndCmd(execGet)
}

func execGet(_ context.Context, _ io.Writer, cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
