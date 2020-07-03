package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
	"io"
)

//Create Command for get pod
func NewCmdGetServiceAccount() *cobra.Command {
	return NewCmd("serviceaccount").
		WithDescription("Get secrets list").
		SetAliases([]string{"sa", "serviceaccounts"}).
		RunWithNoArgs(execGetServiceAccount)
}

// Function for get command
func execGetServiceAccount(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		serviceaccounts, err := getAllRawServiceAccount(ctx, executor.Client, executor.Namespace, NO_STRING)
		if err != nil {
			return err
		}

		if !renderServiceAccountsListInfo(serviceaccounts) {
			color.Red.Fprintln(out, "No secret exists in the namespace")
		}

		return nil
	})
}
