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
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)


//Create Command for get pod
func NewCmdConfig() *cobra.Command {
	return NewCmd("config").
		WithDescription("Manage Kubernetes Config").
		AddConfigGroups().
		RunWithNoArgs(execConfig)
}

//Search Label command
func NewCmdConfigDelete() *cobra.Command {
	return NewCmd("delete").
		WithDescription("Delete specific cluster configuration in kbueconfig").
		SetAliases([]string{"del"}).
		RunWithArgsAndCmd(execDeleteConfig)
}

// Function for Config execution
func execConfig(_ context.Context, _ io.Writer) error {
	return nil
}

// Function for delete configuration in kubeconfig
func execDeleteConfig(ctx context.Context, out io.Writer, cmd *cobra.Command, args []string) error {
	return runExecutor(ctx, func(executor Executor) error {
		configAccess := clientcmd.NewDefaultPathOptions()

		deleteClusterConfig(out, configAccess, cmd)
		return nil
	})
}

// Delete Configuration
func deleteClusterConfig(out io.Writer, configAccess clientcmd.ConfigAccess, cmd *cobra.Command) error {
	var targetContexts []string
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	configFile := configAccess.GetDefaultFilename()
	if configAccess.IsExplicitFile() {
		configFile = configAccess.GetExplicitFile()
	}

	//Select Target
	args := cmd.Flags().Args()
	if len(args) == 0 {
		contextList := []string{}

		// get list of context
		for name, _ := range config.Contexts {
			contextList = append(contextList, name)
		}

		// Choose context to delete
		prompt := &survey.MultiSelect{
			Message: "Pick contexts you want to delete:",
			Options: contextList,
		}
		survey.AskOne(prompt, &targetContexts)

		if len(targetContexts) == 0 {
			color.Red.Fprintln(out, "No context has been selected")
			return nil
		}
	}

	for _, target := range targetContexts {
		cluster := config.Contexts[target].Cluster
		name := config.Contexts[target].AuthInfo

		//Delete cluster
		_, ok := config.Clusters[cluster]
		if !ok {
			color.Red.Fprintln(out, fmt.Sprintf("cannot delete cluster %s, not in %s", cluster, configFile))
		} else {
			delete(config.Clusters, cluster)
		}

		//Delete AuthInfo
		_, ok = config.AuthInfos[name]
		if !ok {
			color.Red.Fprintln(out, fmt.Sprintf("cannot delete auth info %s, not in %s", name, configFile))
		} else {
			delete(config.AuthInfos, name)
		}

		//Delete Context
		_, ok = config.Contexts[target]
		if !ok {
			color.Red.Fprintln(out, fmt.Sprintf("cannot delete context %s, not in %s", target, configFile))
		} else {
			delete(config.Contexts, target)
		}
	}

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return err
	}

	color.Blue.Fprintln(out, fmt.Sprintf("Deleted context %s from %s\n", strings.Join(targetContexts, ","), configFile))

	return nil
}


