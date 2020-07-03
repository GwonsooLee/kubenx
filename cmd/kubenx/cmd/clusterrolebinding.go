package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get cluster role
func NewCmdGetClusterRoleBinding() *cobra.Command {
	return NewCmd("clusterrolebinding").
		WithDescription("Get clusterrole list").
		SetAliases([]string{"clusterrolebindings"}).
		RunWithNoArgs(execGetClusterRoleBinding)
}

// Function for get command
func execGetClusterRoleBinding(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		clusterRoleBindings, err := getAllRawClusterRoleBindings(ctx, executor.RbacV1Client, NO_STRING)
		if err != nil {
			return err
		}

		if !renderClusterRoleBindingsListInfo(clusterRoleBindings) {
			color.Red.Fprintln(out, "No cluster role binding exists in the namespace")
		}

		return nil
	})
}
