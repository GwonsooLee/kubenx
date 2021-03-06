package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get cluster role
func NewCmdGetClusterRole() *cobra.Command {
	return NewCmd("clusterrole").
		WithDescription("Get clusterrole list").
		SetAliases([]string{"cr", "clusterroles"}).
		RunWithNoArgs(execGetClusterrole)
}

// Function for get command
func execGetClusterrole(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		clusterRoles, err := runner.GetAllRawClusterRoles(ctx, executor.RbacV1Client, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderClusterRolesListInfo(clusterRoles) {
			color.Red.Fprintln(out, "No cluster role exists in the namespace")
		}

		return nil
	})
}
