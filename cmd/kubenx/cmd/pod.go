package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

//Create Command for get pod
func NewCmdGetPod() *cobra.Command {
	return NewCmd("pod").
		WithDescription("Get pod list").
		SetAliases([]string{"po", "pods"}).
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

		// Get All Pods in current namespace
		pods, err := runner.GetAllRawPods(ctx, executor.Client, executor.Namespace, utils.NO_STRING)
		if err != nil {
			return err
		}

		if !runner.RenderPodListInfo(pods) {
			color.Red.Fprintln(out, "No pod exists in the namespace")
		}

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
		pod, local_port, pod_port := runner.SelectPodPortNS(podNames)

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
			err := runner.PortForwardToPod(runner.PortForwardAPodRequest{
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
