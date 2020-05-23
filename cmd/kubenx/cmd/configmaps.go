package cmd

import (
	"github.com/GwonsooLee/kubenx/pkg/color"
	"io"
	"context"
	"github.com/spf13/cobra"
)

//Create Command for get pod
func NewCmdGetConfigMap() *cobra.Command {
	return NewCmd("configmap").
		WithDescription("Get configmaps").
		SetAliases([]string{"configmaps", "cm"}).
		RunWithNoArgs(execGetConfigMap)
}


// Function for get command
func execGetConfigMap(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		configMaps, err := getAllRawConfigMaps(ctx, executor.Client, executor.Namespace, NO_STRING)
		if err != nil {
			return err
		}
		if ! renderConfigMapsListInfo(configMaps) {
			color.Red.Fprintln(out, "No configmap exists in the namespace")
		}

		return nil
	})
}

