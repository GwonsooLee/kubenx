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
	"strings"
	"context"
	"encoding/base64"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AlecAivazis/survey/v2"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/GwonsooLee/kubenx/pkg/aws"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"k8s.io/client-go/tools/clientcmd/api"
)


//Create Command for get pod
func NewCmdConfig() *cobra.Command {
	return NewCmd("config").
		WithDescription("Manage Kubernetes Config").
		AddConfigGroups().
		SetFlags().
		RunWithNoArgs(execConfig)
}

//Update config command
func NewCmdConfigUpdate() *cobra.Command {
	return NewCmd("update").
		WithDescription("Update specific cluster configuration in kbueconfig").
		RunWithArgs(execUpdateConfig)
}

//Delete config command
func NewCmdConfigDelete() *cobra.Command {
	return NewCmd("delete").
		WithDescription("Delete specific cluster configuration in kbueconfig").
		SetAliases([]string{"del"}).
		RunWithArgsAndCmd(execDeleteConfig)
}

// Function for Config execution
func execConfig(_ context.Context, out io.Writer) error {
	// Get Current Config
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	color.Blue.Fprintln(out, "[ Current available context ]")
	for context, _ := range config.Contexts {
		color.Green.Fprintln(out, context)
	}

	return nil
}


// Function for update configuration in kubeconfig
func execUpdateConfig(ctx context.Context, out io.Writer, args []string) error {
	return runExecutorWithAWS(ctx, func(executor Executor) error {
		var cluster string

		// 1. Check Cluster
		if len(args) == 1 {
			cluster = args[0]
		}

		if len(cluster) == 0 {
			clusters := getEKSClusterList(executor.EKS)
			prompt := &survey.Select{
				Message: "Choose a cluster:",
				Options: clusters,
			}
			survey.AskOne(prompt, &cluster)
		}

		// if cluster is not set
		if cluster == "" {
			color.Red.Fprintln(out, fmt.Sprintf("No cluster exists in %s region...", viper.GetString("region")))
			return nil
		}

		// Get Current Config
		configAccess := clientcmd.NewDefaultPathOptions()
		config, err := configAccess.GetStartingConfig()
		if err != nil {
			return err
		}

		// 2. Get Cluster Information
		clusterInfo, err := aws.GetClusterInfo(executor.EKS, cluster)
		if err != nil {
			return err
		}

		arn 	:= *clusterInfo.Cluster.Arn
		name 	:= *clusterInfo.Cluster.Name

		//Check existing cluster
		isUpdated := false
		for _, c := range config.Clusters {
			if c.Server ==  *clusterInfo.Cluster.Endpoint {
				color.Red.Fprintln(out, fmt.Sprintf("%s config already exist", name))
				isUpdated = true
			}
		}

		newCluster := api.NewCluster()
		decoded, _ := base64.StdEncoding.DecodeString(*clusterInfo.Cluster.CertificateAuthority.Data)
		newCluster.CertificateAuthorityData = decoded
		newCluster.Server 					= *clusterInfo.Cluster.Endpoint

		newAuthInfo := api.NewAuthInfo()
		newAuthInfo.Exec			= &api.ExecConfig{
			Command:    AUTH_COMMAND,
			Args:       []string{"--region", viper.GetString("region"), "eks", "get-token", "--cluster-name", name},
			APIVersion: AUTH_API_VERSION,
		}

		newContext := api.NewContext()
		newContext.Cluster 					= arn
		newContext.AuthInfo					= arn


		config.Clusters[arn] 	= newCluster
		config.AuthInfos[arn] 	= newAuthInfo
		config.Contexts[name]	= newContext

		if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
			return err
		}

		if isUpdated {
			color.Blue.Fprintln(out, fmt.Sprintf("Update existing context %s", name))
		} else {
			color.Blue.Fprintln(out, fmt.Sprintf("Create new context %s", name))
		}


		return nil
	})
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
	targetContexts := []string{}
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
	}else {
		targetContexts = append(targetContexts, args[0])
	}

	for _, target := range targetContexts {
		//Delete Context
		_, ok := config.Contexts[target]
		if !ok {
			color.Red.Fprintln(out, fmt.Sprintf("cannot delete context %s, not in %s", target, configFile))
			return nil
		}


		cluster := config.Contexts[target].Cluster
		name := config.Contexts[target].AuthInfo

		//Delete cluster
		_, ok = config.Clusters[cluster]
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

		delete(config.Contexts, target)
	}

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return err
	}

	color.Blue.Fprintln(out, fmt.Sprintf("Deleted context %s from %s\n", strings.Join(targetContexts, ","), configFile))

	return nil
}


