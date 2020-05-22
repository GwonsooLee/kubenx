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
	"github.com/spf13/cobra"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/table"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//Create Command for get pod
func NewCmdGetPod() *cobra.Command {
	return NewCmd("pod").
		WithDescription("Get pod list").
		SetAliases([]string{"po", "pods"}).
		SetFlags().
		RunWithNoArgs(execGetPod)
}

// Start Port Forwarding
func NewCmdPortForward() *cobra.Command {
	return NewCmd("port-forward").
		WithDescription("Port Forward Commamd for connecting to pod.").
		SetAliases([]string{"pf"}).
		RunWithNoArgs(execPortForward)
}

// Function for get command
func execGetPod(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		opts := metav1.ListOptions{}

		// Get All Pods in current namespace
		pods, err := executor.Client.CoreV1().Pods(executor.Namespace).List(ctx, opts)
		if err != nil {
			color.Red.Fprintln(out, err)
			os.Exit(1)
		}

		if len(pods.Items) <= 0 {
			color.Red.Fprintln(out, "No pod exists in the namespace")
			return nil
		}

		// Table setup
		table := table.GetTableObject()
		table.SetHeader([]string{"Name", "READY", "STATUS", "HOSTNAME", "POD IP", "HOST IP", "NODE", "AGE"})

		now := time.Now()
		for _, pod := range pods.Items {
			objectMeta := pod.ObjectMeta
			podStatus := pod.Status
			podSpec := pod.Spec
			duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

			readyCount := 0
			totalCount := 0
			var status string

			status = fmt.Sprintf("%v", podStatus.Phase)
			for _, containerStatus := range podStatus.ContainerStatuses {
				totalCount += 1
				if containerStatus.Ready {
					readyCount += 1
				}

				if containerStatus.State.Waiting != nil {
					status = containerStatus.State.Waiting.Reason
					break
				}

				if containerStatus.State.Running != nil {
					status = "Running"
				}

				if containerStatus.State.Terminated != nil {
					status = fmt.Sprintf("%s (Exit Code: %d)", containerStatus.State.Terminated.Reason, containerStatus.State.Terminated.ExitCode)
				}
			}

			table.Append([]string{objectMeta.Name, strconv.Itoa(readyCount) + "/" + strconv.Itoa(totalCount), status, podSpec.Hostname, podStatus.PodIP, podStatus.HostIP, podSpec.NodeName, duration})
		}
		table.Render()
		return nil
	})
}


// Function for port forward
func execPortForward(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		var wg sync.WaitGroup

		wg.Add(1)

		//Get All Pods
		pods, err := executor.Client.CoreV1().Pods(executor.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		// Retrieve only strings from the list
		var podNames []string
		for _, pod := range pods.Items {
			objectMeta := pod.ObjectMeta
			podNames = append(podNames, objectMeta.Name)
		}

		//Get Parameters from input
		pod, local_port, pod_port := selectPodPortNS(podNames)

		// stopCh control the port forwarding lifecycle. When it gets closed the
		// port forward will terminate
		stopCh := make(chan struct{}, 1)
		// readyCh communicate when the port forward is ready to get traffic
		readyCh := make(chan struct{})
		// stream is used to tell the port forwarder where to place its output or
		// where to expect input if needed. For the port forwarding we just need
		// the output eventually
		stream := genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}

		// managing termination signal from the terminal. As you can see the stopCh
		// gets closed to gracefully handle its termination.
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("Finishing port forwarding")
			close(stopCh)
			wg.Done()
		}()

		go func() {
			// PortForward the pod specified from its port 9090 to the local port
			// 8080
			err := _port_forward_to_pod(PortForwardAPodRequest{
				RestConfig: executor.Config,
				Pod: corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pod,
						Namespace: executor.Namespace,
					},
				},
				LocalPort: local_port,
				PodPort:   pod_port,
				Streams:   stream,
				StopCh:    stopCh,
				ReadyCh:   readyCh,
			})
			if err != nil {
				panic(err)
			}
		}()

		select {
		case <-readyCh:
			break
		}
		println("Port forwarding is ready to get traffic. have fun!")

		wg.Wait()

		return err
	})
}
