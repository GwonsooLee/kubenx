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
func NewCmdGetSecret() *cobra.Command {
	return NewCmd("secret").
		WithDescription("Get secrets list").
		SetAliases([]string{"secrets"}).
		RunWithNoArgs(execGetSecret)
}

// Function for get command
func execGetSecret(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		// Get All Pods in current namespace
		secrets, err := runner.GetAllRawSecrets(ctx, executor.Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderSecretsListInfo(secrets) {
			color.Red.Fprintln(out, "No secret exists in the namespace")
		}

		return nil
	})
}
