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
		roles, err := runner.GetAllRawRoles(ctx, executor.RbacV1Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderRolesListInfo(roles) {
			color.Red.Fprintln(out, "No role exists in the namespace")
		}

		return nil
	})
}
