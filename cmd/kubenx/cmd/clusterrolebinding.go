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
		clusterRoleBindings, err := runner.GetAllRawClusterRoleBindings(ctx, executor.RbacV1Client, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderClusterRoleBindingsListInfo(clusterRoleBindings) {
			color.Red.Fprintln(out, "No cluster role binding exists in the namespace")
		}

		return nil
	})
}
