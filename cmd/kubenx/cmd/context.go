package cmd

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/aws"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/tools/clientcmd"
	"os/exec"
)

//Create Command for get pod
func NewCmdContext() *cobra.Command {
	return NewCmd("context").
		WithDescription("Change context from kubeconfig").
		SetAliases([]string{"ctx"}).
		RunWithArgs(execContext)
}

// Function for changing context
func execContext(_ context.Context, out io.Writer, args []string) error {
	return changeContext(out, args)
}

//Change Context
func changeContext(out io.Writer, args []string) error {
	var contextList []string
	var newContext string

	//Get API configuration
	configs, _, err := runner.GetAPIConfig()
	if err != nil {
		color.Red.Fprintln(out, err.Error())
		return err
	}

	// Get Client Configuration
	currentConfig, err := runner.GetCurrentConfig()
	if err != nil {
		color.Red.Fprintln(out, err.Error())
		return err
	}

	//getting current Context
	currentContext := currentConfig.CurrentContext

	if len(args) == 0 {
		// get list of context
		for context, _ := range configs.Contexts {
			contextList = append(contextList, context)
		}

		// Get New Context
		color.Red.Fprintln(out, fmt.Sprintf("Current Context: %s", currentContext))
		prompt := &survey.Select{
			Message: "Choose Context:",
			Options: contextList,
		}
		survey.AskOne(prompt, &newContext)

		if newContext == "" {
			color.Red.Fprintln(out, fmt.Errorf("Changing Context has been canceled"))
			return err
		}
	} else {
		newContext = args[0]
	}

	//Change To New Context
	currentConfig.CurrentContext = newContext
	configAccess := clientcmd.NewDefaultClientConfig(*configs, &clientcmd.ConfigOverrides{}).ConfigAccess()

	clientcmd.ModifyConfig(configAccess, *currentConfig, false)
	color.Yellow.Fprintln(out, fmt.Sprintf("Context is changed to %s", newContext))

	//Assume Role
	kubeEKSConfig, err := aws.FindEKSAussmeInfo()
	if err != nil {
		return err
	}

	if _, ok := kubeEKSConfig.EKSAssumeMapping[newContext]; !ok {
		color.Red.Fprintln(out, "Assume Information is not mapped to $HOME/.kubenx/config .")
		return nil
	}

	assumeCreds := aws.AssumeRole(kubeEKSConfig.Assume[kubeEKSConfig.EKSAssumeMapping[newContext]], kubeEKSConfig.SessionName)

	pbcopy := exec.Command("pbcopy")
	in, _ := pbcopy.StdinPipe()

	if err := pbcopy.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s\n", *assumeCreds.AccessKeyId))); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_SECRET_ACCESS_KEY=%s\n", *assumeCreds.SecretAccessKey))); err != nil {
		return err
	}

	if _, err := in.Write([]byte(fmt.Sprintf("export AWS_SESSION_TOKEN=%s\n", *assumeCreds.SessionToken))); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	err = pbcopy.Wait()
	if err != nil {
		color.Red.Fprintln(out, err.Error())
		return err
	}

	color.Blue.Fprintln(out, "Assume Credentials copied to clipboard, please paste it.")

	return nil
}
