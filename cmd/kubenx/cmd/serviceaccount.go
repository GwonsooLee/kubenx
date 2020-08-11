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
func NewCmdGetServiceAccount() *cobra.Command {
	return NewCmd("serviceaccount").
		WithDescription("Get service account list").
		SetAliases([]string{"sa", "serviceaccounts"}).
		RunWithNoArgs(execGetServiceAccount)
}

// Function for get command
func execGetServiceAccount(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		serviceAccounts, err := runner.GetAllRawServiceAccount(ctx, executor.Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderServiceAccountsListInfo(serviceAccounts) {
			color.Red.Fprintln(out, "No secret exists in the namespace")
		}

		return nil
	})
}
