package cmd

import (
	"context"
	"fmt"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get pod
func NewCmdVersion() *cobra.Command {
	return NewCmd("version").
		WithDescription("Find kubenx release version").
		RunWithNoArgs(execVersion)
}

// Function for search execution
func execVersion(_ context.Context, out io.Writer) error {
	version := "v1.0.0"
	color.Blue.Fprintln(out, fmt.Sprintf("Current Version is %s\n", version))
	return nil
}
