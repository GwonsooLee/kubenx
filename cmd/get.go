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
	"strings"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
	"github.com/olekukonko/tablewriter"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get detailed information about eks cluster",
	Long: `Get detailed information about eks cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		//Get Arguments
		argsLen := len(args)

		if argsLen > 1 {
			red("Too many Arguments")
			os.Exit(1)
		} else if argsLen == 0 {
			red("At least one argument is needed.")
			os.Exit(1)
		}

		objType := args[0]

		// Call function according to the second third parameter
		switch {
		case objType == "cluster":
			_get_detail_info_of_cluster()
		default:
			red("Please follow the direction")
		}

	},
}

var (
	TAG_PREFIX = "kubernetes"

)

func init() {
	//Local Option
	getCmd.Flags().StringP("cluster", "c", "", "Name of EKS Cluster")

	//Bind Flag to viper
	viper.BindPFlag("cluster", getCmd.Flags().Lookup("cluster"))
	rootCmd.AddCommand(getCmd)
}

// Get detail information about cluster
func _get_detail_info_of_cluster()  {
	svc := _get_eks_session()

	// Check the cluster First
	cluster := ""
	cluster = viper.GetString("cluster")

	// If cluster is not given then choose!
	if len(cluster) <= 0 {
		cluster = _choose_cluster()
	}


	// 1. Get Cluster Information
	clusterInfo := _get_cluster_info_with_session(svc, cluster)

	clusterTable := tablewriter.NewWriter(os.Stdout)
	clusterTable.SetHeader([]string{"Name", cluster})
	clusterTable.Append([]string{"Version", *clusterInfo.Cluster.Version})
	clusterTable.Append([]string{"Status", *clusterInfo.Cluster.Status})
	clusterTable.Append([]string{"Arn", *clusterInfo.Cluster.Arn})
	clusterTable.Append([]string{"Endpoint", *clusterInfo.Cluster.Endpoint})
	clusterTable.Append([]string{"Cluster SG", *clusterInfo.Cluster.ResourcesVpcConfig.ClusterSecurityGroupId})

	// Get VPC Information
	ec2_svc := _get_ec2_session()
	vpcInfo := _get_vpc_info(ec2_svc, clusterInfo.Cluster.ResourcesVpcConfig.VpcId)
	vpcName := ""
	for _, obj := range vpcInfo.Vpcs[0].Tags {
		if *obj.Key == "Name" {
			vpcName = *obj.Value
			break
		}
	}
	vpcStr := vpcName + "(" + *vpcInfo.Vpcs[0].VpcId + ")"
	clusterTable.Append([]string{"VPC ID", vpcStr})
	clusterTable.Append([]string{"VPC Cidr Block", *vpcInfo.Vpcs[0].CidrBlock})

	// Get Subnet Information
	subnetList := clusterInfo.Cluster.ResourcesVpcConfig.SubnetIds
	if len(subnetList) > 0 {
		var subnetIds []*string
		for _, subnetId := range subnetList {
			subnetIds = append(subnetIds, subnetId)
		}

		subnetInfo := _get_subnet_info(ec2_svc, subnetIds)
		for _, subnetObj := range subnetInfo.Subnets {
			az := subnetObj.AvailabilityZone
			//subnetId := subnetObj.SubnetId
			tags := ""
			subnetName := ""
			for _, obj := range subnetObj.Tags {
				if strings.HasPrefix(*obj.Key, TAG_PREFIX){
					line := *obj.Key+"="+*obj.Value+" "
					tags += line
				} else if *obj.Key == "Name" {
					subnetName = *obj.Value
				}
			}

			clusterTable.Append([]string{subnetName+"("+*az+")", tags})
		}
	}
	clusterTable.SetAlignment(tablewriter.ALIGN_LEFT)
	clusterTable.Render()
}
