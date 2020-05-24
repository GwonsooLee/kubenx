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
		RunWithArgsAndCmd(execSearch)
}

//Search Label command
func NewCmdSearchLabel() *cobra.Command {
	return NewCmd("label").
		WithDescription("Search resources by label").
		SetAliases([]string{"la", "lab"}).
		RunWithNoArgs(execSearchLabel)
}

// Function for search execution
func execSearch(_ context.Context, _ io.Writer, cmd *cobra.Command, args []string) error {
	cmd.Help()
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
		pods, err := getAllRawPods(ctx, executor.Client, ALL_NAMESPACE, labelSelector)
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


