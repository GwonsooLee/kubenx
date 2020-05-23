/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io"
	"fmt"
	"context"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/color"
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
	configs, _, err := getAPIConfig()
	if err != nil {
		color.Red.Fprintln(out, err.Error())
		return err
	}

	// Get Client Configuration
	currentConfig, err := getCurrentConfig()
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

	return nil
}
