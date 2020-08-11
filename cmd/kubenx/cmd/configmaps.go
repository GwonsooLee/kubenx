package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"io"
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
		configMaps, err := runner.GetAllRawConfigMaps(ctx, executor.Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}
		if !runner.RenderConfigMapsListInfo(configMaps) {
			color.Red.Fprintln(out, "No configmap exists in the namespace")
		}

		return nil
	})
}
