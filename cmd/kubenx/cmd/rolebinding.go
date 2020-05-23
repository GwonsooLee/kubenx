package cmd

import (
	"github.com/GwonsooLee/kubenx/pkg/color"
	"io"
	"context"
	"github.com/spf13/cobra"
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
		roles, err := getAllRawRoleBindings(ctx, executor.RbacV1Client, executor.Namespace, NO_STRING)
		if err != nil {
			return err
		}

		if ! renderRoleBindingsListInfo(roles) {
			color.Red.Fprintln(out, "No rolebinding exists in the namespace")
		}

		return nil
	})
}

