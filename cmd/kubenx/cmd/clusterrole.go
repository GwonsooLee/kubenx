package cmd

import (
	"github.com/GwonsooLee/kubenx/pkg/color"
	"io"
	"context"
	"github.com/spf13/cobra"
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
		clusterRoles, err := getAllRawClusterRoles(ctx, executor.RbacV1Client, NO_STRING)
		if err != nil {
			return err
		}

		if ! renderClusterRolesListInfo(clusterRoles) {
			color.Red.Fprintln(out, "No cluster role exists in the namespace")
		}

		return nil
	})
}

