package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get pod
func NewCmdGetRole() *cobra.Command {
	return NewCmd("role").
		WithDescription("Get role list").
		SetAliases([]string{"roles"}).
		RunWithNoArgs(execGetRole)
}

// Function for get command
func execGetRole(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		roles, err := getAllRawRoles(ctx, executor.RbacV1Client, executor.Namespace, NO_STRING)
		if err != nil {
			return err
		}

		if !renderRolesListInfo(roles) {
			color.Red.Fprintln(out, "No role exists in the namespace")
		}

		return nil
	})
}
