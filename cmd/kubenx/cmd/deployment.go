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
	"time"
	"context"
	"github.com/spf13/cobra"
	"github.com/GwonsooLee/kubenx/pkg/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/GwonsooLee/kubenx/pkg/utils"
)

//Create Command for get service
func NewCmdGetDeployment() *cobra.Command {
	return NewCmd("deployment").
		WithDescription("Retrieve information about deployment in detail").
		SetAliases([]string{"dep", "deploy"}).
		RunWithNoArgs(execGetDeployment)
}

// Function for getting services
func execGetDeployment(ctx context.Context, out io.Writer) error {

	return runExecutor(ctx, func(executor Executor) error {
		now := time.Now()
		var depList []string

		//Get all deployments list in the namespace
		errorCount := 0

		//Tables to show information
		table := table.GetTableObject()
		table.SetHeader([]string{"Name", "READY", "UP-TO-DATE", "AVAILABLE", "STRATEGY TYPE", "MAX UNAVAILABLE", "NAX SURGE", "CONTAINERS", "IMAGE", "AGE"})

		// Search Deployment with v1beta1
		deployments, err := executor.BetaV1Client.Deployments(executor.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorCount += 1
		} else {
			//Start Searching
			for _, deployment := range deployments.Items {
				//Get Object Meta Data
				objectMeta := deployment.ObjectMeta
				spec := deployment.Spec

				duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

				// Check if it is using RollingUpdate strategy in order to get MaxUnAvailable, MaxSurge
				maxUnavailable := ""
				maxSurge := ""
				if spec.Strategy.RollingUpdate != nil {
					maxUnavailable = spec.Strategy.RollingUpdate.MaxUnavailable.StrVal
					maxSurge = spec.Strategy.RollingUpdate.MaxSurge.StrVal
				}

				// Get container spec in pod
				podSpec := spec.Template.Spec
				nameString := ""
				imageString := ""
				for _, container := range podSpec.Containers {
					nameString += container.Name + "\n"
					imageString += utils.RemoveSHATags(container.Image) + "\n"
				}

				table.Append([]string{objectMeta.Name, utils.Int32ToString(*spec.Replicas), utils.Int32ToString(deployment.Status.UpdatedReplicas), utils.Int32ToString(deployment.Status.AvailableReplicas), string(spec.Strategy.Type), maxUnavailable, maxSurge, nameString, imageString, duration})
				depList = append(depList, objectMeta.Name)
			}
		}

		// Get Deployment with core v1 version
		deploymentsV1, err := executor.Client.AppsV1().Deployments(executor.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorCount += 1
		} else {
			//Start Searching
			for _, deployment := range deploymentsV1.Items {
				//Get Object Meta Data
				objectMeta := deployment.ObjectMeta

				hasValue := false
				for _, element := range depList {
					if element == objectMeta.Name {
						hasValue = true
						break
					}
				}

				if hasValue { continue}

				spec := deployment.Spec
				duration := duration.HumanDuration(now.Sub(objectMeta.CreationTimestamp.Time))

				// Check if it is using RollingUpdate strategy in order to get MaxUnAvailable, MaxSurge
				maxUnavailable := ""
				maxSurge := ""
				if spec.Strategy.RollingUpdate != nil {
					maxUnavailable = spec.Strategy.RollingUpdate.MaxUnavailable.StrVal
					maxSurge = spec.Strategy.RollingUpdate.MaxSurge.StrVal
				}

				// Get container spec in pod
				podSpec := spec.Template.Spec
				nameString := ""
				imageString := ""
				for _, container := range podSpec.Containers {
					nameString += container.Name + "\n"
					imageString += utils.RemoveSHATags(container.Image) + "\n"
				}

				table.Append([]string{objectMeta.Name, utils.Int32ToString(*spec.Replicas), utils.Int32ToString(deployment.Status.UpdatedReplicas), utils.Int32ToString(deployment.Status.AvailableReplicas), string(spec.Strategy.Type), maxUnavailable, maxSurge, nameString, imageString, duration})
			}
		}

		// If Error count is larger than 1, which means no deployment int the cluster
		if errorCount > 1 {
			color.Red.Fprintln(out,"The server could not find the requested resource")
			return err
		}
		table.Render()
		return nil
	})
}

