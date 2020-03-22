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

	"github.com/spf13/cobra"
)

// podCmd represents the pod command
var podCmd = &cobra.Command{
	Use:   "pod",
	Short: "Customized command about pod",
	Long: `Customized command about pod`,
	Run: func(cmd *cobra.Command, args []string) {

		argsLen := len(args)

		if argsLen > 1 {
			red("Too many Arguments")
			os.Exit(1)
		} else if argsLen == 0 {
			_get_pod_list()
		} else {
			//objType := args[0]

			// Call function according to the second third parameter
			//switch {
			//case objType == "cluster":
			//	_get_detail_info_of_cluster()
			//case objType == "nodegroup" || objType == "ng":
			//	_get_detail_info_of_nodegroup()
			//case objType == "securitygroup" || objType == "sg":
			//	_get_detail_info_of_security_group()
			//default:
			//	red("Please follow the direction")
			//}
		}
	},
}

func init() {
	rootCmd.AddCommand(podCmd)
}
