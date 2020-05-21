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
	"github.com/spf13/cobra"
	"os"
)

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Command about namespace",
	Long:  `Command about namespace`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			Red("Too many Arguments")
			os.Exit(1)
		} else if len(args) == 1 {
			_change_current_namespace(args[0])
		} else {
			_change_current_namespace("")
		}
	},
	Aliases: []string{"ns"},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
}
