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
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AlecAivazis/survey/v2"
)
// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List kubernetes resources",
	Long: `List kubernetes resources`,
	Run: func(cmd *cobra.Command, args []string) {
		//Get Arguments
		argsLen := len(args)

		if argsLen > 1 {
			red("Too many Arguments")
			os.Exit(1)
		}

		objType := args[0]

		switch {
		case objType == "cluster":
			list_clusters()
		case objType == "nodegroup" || objType == "ng":
			list_nodegroups()
		default:
			red("Please follow the direction")
		}


	},
}

func init() {
	// Get Flag Values
	listCmd.Flags().StringP("cluster", "c", "", "EKS Cluster Name")

	//Bind Flag to viper
	viper.BindPFlag("cluster", listCmd.Flags().Lookup("cluster"))

	rootCmd.AddCommand(listCmd)
}

// List Clusters
func list_clusters()  {
	svc := get_eks_session()
	clusters := get_eks_clusters(svc)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Version", "Arn", "Endpoint"})
	for _, cluster := range clusters {
		inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(cluster)}

		clusterInfo, err := svc.DescribeCluster(inputParamsDesc)
		if err != nil {
			red(err)
			os.Exit(1)
		}

		arn := clusterInfo.Cluster.Arn
		endpoint := clusterInfo.Cluster.Endpoint
		version := clusterInfo.Cluster.Version
		status := clusterInfo.Cluster.Status

		//tableData = append(tableData, []string{*cluster, *status, *version, *arn, *endpoint})
		table.Append([]string{cluster, *status, *version, *arn, *endpoint})
	}

	if len(clusters) == 0 {
		table.Append([]string{"No Cluster exists", "", "", "", ""})
		table.SetAutoMergeCells(true)
	}
	table.Render()
}

// List nodeGroups
func list_nodegroups() {
	// Check the cluster First
	cluster := ""
	cluster = viper.GetString("cluster")

	// If cluster is not given then choose!
	if len(cluster) <= 0 {
		options := get_eks_clusters(nil)
		prompt := &survey.Select{
			Message: "Choose a color:",
			Options: options,
		}
		survey.AskOne(prompt, &cluster)
	}

	svc := get_eks_session()

	inputParams := &eks.ListNodegroupsInput{ClusterName: aws.String(cluster), MaxResults: aws.Int64(100)}
	res, err := svc.ListNodegroups(inputParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	// Tables for showing outputs
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "STATUS", "INSTANCE TYPE", "LABELS", "MIN SIZE", "DISIRED SIZE", "MAX SIZE", "AUTOSCALING GROUPDS", "DISK SIZE"})
	// Get node group information
	for _, nodegroup := range res.Nodegroups {
		ngParams := &eks.DescribeNodegroupInput{ClusterName: aws.String(cluster), NodegroupName: aws.String(*nodegroup)}
		info, err := svc.DescribeNodegroup(ngParams)
		if err != nil {
			red(err)
			os.Exit(1)
		}

		// Retrieve Values
		//
		//Status string
		Status := info.Nodegroup.Status

		// InstanceTypes Array
		InstanceTypesArray := info.Nodegroup.InstanceTypes
		instanceTypes := ""
		for _, iType := range InstanceTypesArray {
			line := *iType+","
			instanceTypes += line
		}
		instanceTypes = instanceTypes[:len(instanceTypes)-1]

		// Label Map(Object)
		LabelsObj := info.Nodegroup.Labels
		labels := ""
		for key, value := range LabelsObj {
			line := key+"="+*value+","
			labels += line
		}
		labels = labels[:len(labels)-1]

		// InstanceTypes Array of Map(Object)
		AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
		autoScalingGroups := ""
		for _, group := range AutoScalingGroupsArray {
			line := *(group.Name)+","
			autoScalingGroups += line
		}
		autoScalingGroups = autoScalingGroups[:len(autoScalingGroups)-1]

		// Desired, Min, Max Size ==> int64
		DesiredSize := info.Nodegroup.ScalingConfig.DesiredSize
		MaxSize := info.Nodegroup.ScalingConfig.MaxSize
		MinSize := info.Nodegroup.ScalingConfig.MinSize

		// DiskSize int64
		DiskSize := info.Nodegroup.DiskSize

		table.Append([]string{*nodegroup, *Status, instanceTypes, labels, strconv.FormatInt(*MinSize, 10), strconv.FormatInt(*DesiredSize, 10), strconv.FormatInt(*MaxSize, 10), autoScalingGroups, strconv.FormatInt(*DiskSize, 10)})
	}
	table.Render()
}


// Get EKS Session
func get_eks_session() *eks.EKS {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Get All EKS Cluster
func get_eks_clusters(svc *eks.EKS) []string {
	if(svc == nil){
		svc = get_eks_session()
	}

	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	var ret []string
	for _, cluster := range res.Clusters {
		ret = append(ret, *cluster)
	}

	return ret
}
