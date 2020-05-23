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
	"context"
	"github.com/spf13/cobra"
)

//Create Command for get pod
func NewCmdInspect() *cobra.Command {
	return NewCmd("inspect").
		WithDescription("Inspect resource in detail").
		AddSearchGroups().
		SetAliases([]string{"ins"}).
		AddInspectGroups().
		RunWithNoArgs(execInsepct)
}

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "A brief description of your command",
	Long: `Inspect resource in detail`,
	Run: func(cmd *cobra.Command, args []string) {},
	Aliases: []string{"ins"},
}


// Function for search execution
func execInsepct(_ context.Context, out io.Writer) error {
	return nil
}
