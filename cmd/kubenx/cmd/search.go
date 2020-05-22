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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search resources",
	Long: `Search resources`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			Red("Usage: kubenx search [option]")
			os.Exit(1)
		}

		if args[0] == "label" {
			_search_by_label()
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func _search_by_label()  {
	//Get kubernetes Client
	clientset := _get_k8s_client()

	//namespace
	namespace := _get_namespace()

	labelSelector := _make_label_selector()
	Blue(fmt.Sprintf("Search Selector : %s", labelSelector))
	fmt.Println()

	//Print pod
	Yellow("========Node INFO=======")
	nodes := _get_all_raw_node(clientset, labelSelector)
	_render_node_list_info(nodes)

	//Print pod
	Yellow("========POD INFO=======")
	pods := _get_all_raw_pods(clientset, namespace, labelSelector)
	_render_pod_list_info(pods)

}
