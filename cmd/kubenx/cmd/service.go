package cmd

import (
	"fmt"
	"io"
	"time"
	"strconv"
	"strings"
	"context"
	"github.com/spf13/cobra"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/table"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)


//Create Command for get service
func NewCmdGetService() *cobra.Command {
	return NewCmd("service").
		WithDescription("Get service list").
		SetAliases([]string{"svc", "services"}).
		RunWithNoArgs(execGetService)
}

// Function for getting services
func execGetService(ctx context.Context, out io.Writer) error {

	return runExecutor(ctx, func(executor Executor) error {
		//Get All Pods
		services, err := executor.Client.CoreV1().Services(executor.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			color.Red.Fprintln(out, err.Error())
			return err
		}

		table := table.GetTableObject()
		table.SetHeader([]string{"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "ENDPOINT(S)", "AGE"})

		//Get detailed information about Service
		now := time.Now()
		for _, service := range services.Items {
			objectMeta := service.ObjectMeta

			duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

			serviceSpec := service.Spec
			serviceType := string(serviceSpec.Type)
			ExternalIPs := serviceSpec.ExternalIPs

			// Get Load Balancer Name for this
			externalIP := ""
			if serviceType == "LoadBalancer" {
				externalIP = service.Status.LoadBalancer.Ingress[0].Hostname
			} else if len(ExternalIPs) == 0 {
				externalIP = "<None>"
			}

			// Get Listening Ports of Service
			ports := []string{}
			for _, port := range serviceSpec.Ports {
				appPort := strconv.FormatInt(int64(port.Port), 10)
				ports = append(ports, appPort)

				if port.NodePort != 0 {
					ports = append(ports, fmt.Sprintf("%s/TCP", utils.Int32ToString(port.NodePort)))
				} else {
					ports = append(ports, "/TCP")
				}
			}

			// Convert Label Selector to string set
			labelSelector := []string{}
			if len(serviceSpec.Selector) > 0 {
				for key, value := range serviceSpec.Selector {
					labelSelector = append(labelSelector, key + "=" + value)
				}
			}

			//Get All Pods with endpoint
			pods, err := executor.Client.CoreV1().Pods(executor.Namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(labelSelector,",")})
			if err != nil {
				color.Red.Fprintln(out, err.Error())
				return err
			}

			//Get Endpoints (Pods)
			endpoints := []string{}
			if len(pods.Items) > 0 {
				for _, pod := range pods.Items {
					podIP := pod.Status.PodIP
					endpoints = append(endpoints, podIP)
				}
			}

			table.Append([]string{objectMeta.Name, serviceType, serviceSpec.ClusterIP, externalIP, strings.Join(ports, ":"), strings.Join(endpoints, ","), duration})
		}
		table.Render()
		return nil
	})
}

