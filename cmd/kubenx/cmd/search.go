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
	"os"
	"io"
	"fmt"
	"context"
	"github.com/spf13/cobra"
	"github.com/GwonsooLee/kubenx/pkg/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


//Create Command for get pod
func NewCmdSearch() *cobra.Command {
	return NewCmd("search").
		WithDescription("Search resources").
		AddSearchGroups().
		RunWithNoArgs(execSearch)
}

//Search Label command
func NewCmdSearchLabel() *cobra.Command {
	return NewCmd("label").
		WithDescription("Search resources by label").
		SetAliases([]string{"la", "lab"}).
		RunWithNoArgs(execSearchLabel)
}

// Function for search execution
func execSearch(_ context.Context, out io.Writer) error {
	return nil
}


// Function for search via label
func execSearchLabel(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {

		key, err := getSingleStringInput("Key")
		if err != nil {
			color.Red.Fprintln(out, err)
			os.Exit(1)
		}

		value, err := getSingleStringInput("Value")
		if err != nil {
			color.Red.Fprintln(out, err)
			os.Exit(1)
		}

		labelSelector := fmt.Sprintf("%s=%s", key, value)

		color.Blue.Fprintln(out, fmt.Sprintf("Search Selector : %s", labelSelector))
		fmt.Println()

		//Print pod
		color.Yellow.Fprintln(out, "========Node INFO=======")
		listOpt := metav1.ListOptions{}
		if len(labelSelector) > 0 {
			listOpt = metav1.ListOptions{LabelSelector: labelSelector}
		}

		nodes, err := executor.Client.CoreV1().Nodes().List(ctx, listOpt)
		if err != nil {
			color.Red.Fprintln(out, err)
			os.Exit(1)
		}

		if ! renderNodeListInfo(nodes.Items) {
			color.Red.Fprintln(out, "No node exists")
		}
		fmt.Println()

		//Print pod
		color.Yellow.Fprintln(out, "========Pod INFO=======")
		pods, err := getAllRawPods(executor.Client, ALL_NAMESPACE, labelSelector)
		if err != nil {
			color.Red.Fprintln(out, err)
			os.Exit(1)
		}

		if ! renderPodListInfo(pods) {
			color.Red.Fprintln(out, "No pod exists in the namespace")
		}

		return nil
	})
}


