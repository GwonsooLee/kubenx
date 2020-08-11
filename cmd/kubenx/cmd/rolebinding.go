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
func NewCmdGetRolebinding() *cobra.Command {
	return NewCmd("rolebinding").
		WithDescription("Get role binding list").
		SetAliases([]string{"rolebindings"}).
		RunWithNoArgs(execGetRoleBinding)
}

// Function for get command
func execGetRoleBinding(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		roles, err := runner.GetAllRawRoleBindings(ctx, executor.RbacV1Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderRoleBindingsListInfo(roles) {
			color.Red.Fprintln(out, "No rolebinding exists in the namespace")
		}

		return nil
	})
}
