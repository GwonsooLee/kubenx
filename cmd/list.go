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

	//"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/olekukonko/tablewriter"

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
			list_cluster()
		default:
			red("Please follow the direction")
		}


	},
}

func init() {
	listCmd.Flags().StringP("config", "c", "", "Nothing")
	viper.BindPFlag("config", listCmd.Flags().Lookup("config"))
	rootCmd.AddCommand(listCmd)
}

// List Clusters
func list_cluster()  {
	mySession := session.Must(session.NewSession())
	awsRegion := viper.GetString("region")

	svc := eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	clusters := res.Clusters

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Version", "Arn", "Endpoint"})
	for _, cluster := range clusters {

		inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(*cluster)}

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
		table.Append([]string{*cluster, *status, *version, *arn, *endpoint})

	}

	if len(clusters) == 0 {
		table.Append([]string{"No Cluster exists", "", "", "", ""})
		table.SetAutoMergeCells(true)
	}
	table.Render()
}