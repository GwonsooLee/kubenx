package cmd

import (
	"context"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/table"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"strings"
	"time"
)

//Create Command for get service
func NewCmdGetIngress() *cobra.Command {
	return NewCmd("ingress").
		WithDescription("Retrieve information about ingress").
		SetAliases([]string{"in"}).
		RunWithNoArgs(execGetIngress)
}

// Function for getting services
func execGetIngress(ctx context.Context, out io.Writer) error {
	return runExecutor(ctx, func(executor Executor) error {
		//Get all ingress list in the namespace
		ingresses, err := executor.BetaV1Client.Ingresses(executor.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			if err != nil {
				color.Red.Fprintln(out, err.Error())
				return err
			}
		}

		//Tables to show information
		table := table.GetTableObject()
		table.SetHeader([]string{"Name", "HOST", "ADDRESS", "PATH", "PORTS", "TARGET SERVICE", "AGE"})

		now := time.Now()
		for _, ingress := range ingresses.Items {
			objectMeta := ingress.ObjectMeta
			ingressSpec := ingress.Spec

			var address string
			if len(ingress.Status.LoadBalancer.Ingress) == 0 {
				address = ""
			} else {
				address = ingress.Status.LoadBalancer.Ingress[0].Hostname
			}
			duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

			host := ingressSpec.Rules[0].Host
			if len(host) <= 0 {
				host = "*"
			}

			port := []string{}
			service := []string{}
			paths := []string{}
			for _, path := range ingressSpec.Rules[0].IngressRuleValue.HTTP.Paths {
				if path.Backend.ServicePort.IntVal != 0 {
					port = append(port, utils.Int32ToString(path.Backend.ServicePort.IntVal))
				}
				service = append(service, path.Backend.ServiceName)
				paths = append(paths, path.Path)
			}
			table.Append([]string{objectMeta.Name, host, address, strings.Join(paths, ","), strings.Join(port, ","), strings.Join(service, ","), duration})
		}
		table.Render()
		return nil
	})
}
