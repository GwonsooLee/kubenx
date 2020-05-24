package cmd

import (
	"fmt"
	"io"
	"context"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"github.com/GwonsooLee/kubenx/pkg/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


//Create Command for get pod
func NewCmdGetNode() *cobra.Command {
	return NewCmd("node").
		WithDescription("Get node list").
		SetAliases([]string{"nodes"}).
		RunWithNoArgs(execGetNode)
}


// Function for get command
func execGetNode(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		listOpt := metav1.ListOptions{}
		nodes, err := executor.Client.CoreV1().Nodes().List(context.Background(), listOpt)
		if err != nil {
			return err
		}

		if len(nodes.Items) <= 0 {
			color.Red.Fprintln(out, "No node exists in the namespace")
			return nil
		}


		renderNodeListInfo(nodes.Items)
		return err
	})
}


//Create Command for get pod
func NewCmdInspectNode() *cobra.Command {
	return NewCmd("node").
		WithDescription("Inspect node in detail").
		SetAliases([]string{"nodes"}).
		RunWithNoArgs(execInspectNode)
}

// Function for inspect node command
func execInspectNode(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		//get target node
		target, err := getTargetNode(executor.Client, []string{})
		if err != nil {
			return err
		}

		// Get node information
		detail, err := executor.Client.CoreV1().Nodes().Get(ctx, target, metav1.GetOptions{})
		if err != nil {
			return err
		}

		taints := detail.Spec.Taints

		color.Yellow.Fprintln(out,"========Taint INFO=======")
		for _, taint := range taints {
			txt := fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect)
			Blue(txt)
		}

		if len(taints) == 0 {
			color.Red.Fprintln(out, "There is no taints applied")
		}

		//Get all pods
		pods, _:= getAllRawPods(ctx, executor.Client, executor.Namespace, NO_STRING)

		filtered := []corev1.Pod{}
		for _, pod := range pods {
			if pod.Spec.NodeName == target {
				filtered = append(filtered, pod)
			}
		}

		fmt.Println()
		color.Yellow.Fprintln(out, "========POD INFO=======")
		renderPodListInfo(filtered)

		return nil
	})
}
