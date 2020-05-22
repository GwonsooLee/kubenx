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
	"io"
	"os"
	"github.com/spf13/cobra"
)

func NewCmdGetPod() *cobra.Command {
	return NewCmd("pod").
			WithDescription("Get pod list").
			SetAliases([]string{"po", "pods"}).
			RunWithNoArgs(execGetPod)

}

func execGetPod(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {

	})
}

// podCmd represents the pod command
var getPodCmd = &cobra.Command{
	Use:   "pod",
	Short: "Customized command about pod",
	Long:  `Customized command about pod`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 2 {
			Red("Too many Arguments")
			os.Exit(1)
		}

	},
	Aliases: []string{"po", "pods"},
}

//Port Forward Command
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Port Forward Commamd for connecting to pod",
	Long:  `Port Forward Commamd for connecting to pod`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 2 {
			Red("Too many Arguments")
			os.Exit(1)
		}
		_start_port_forwarding()
	},
	Aliases: []string{"pf"},
}

