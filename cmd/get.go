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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
	"github.com/olekukonko/tablewriter"

	//"github.com/aws/aws-sdk-go/service/ec2"
	//"github.com/aws/aws-sdk-go/aws"
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
		case objType == "nodegroup" || objType == "ng":
			_get_detail_info_of_nodegroup()
		case objType == "securitygroup" || objType == "sg":
			_get_detail_info_of_security_group()
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
	getCmd.Flags().StringP("securitygroup", "s", "", "Name of Security Group")

	//Bind Flag to viper
	viper.BindPFlag("cluster", getCmd.Flags().Lookup("cluster"))
	viper.BindPFlag("securitygroup", getCmd.Flags().Lookup("securitygroup"))
	rootCmd.AddCommand(getCmd)
}

// Get detail information about cluster
func _get_detail_info_of_cluster()  {
	svc := _get_eks_session()

	// Check the cluster First
	cluster := _choose_cluster()


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

// Get Detailed Information about nodegroup
func _get_detail_info_of_nodegroup()  {
	svc := _get_eks_session()

	// Check the cluster First
	cluster := _choose_cluster()

	//Choose Nodegroup
	nodegroup := _choose_nodegroup(cluster)

	// NodeGroup Information Table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "STATUS", "INSTANCE TYPE", "LABELS", "MIN SIZE", "DISIRED SIZE", "MAX SIZE", "AUTOSCALING GROUPDS", "DISK SIZE"})

	// Instance Group Information Table
	instanceTable := tablewriter.NewWriter(os.Stdout)
	instanceTable.SetHeader([]string{"Autoscaling Group", "Instance ID", "Health Status", "Instance Type", "Availability Zone"})

	info := _get_nodegroup_info_with_session(svc, cluster, nodegroup)

	// Retrieve Values
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
	if (len(LabelsObj) > 0 ) {
		for key, value := range LabelsObj {
			line := key+"="+*value+","
			labels += line
		}
		labels = labels[:len(labels)-1]
	}

	// InstanceTypes Array of Map(Object)
	AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
	autoScalingGroups := ""
	for _, group := range AutoScalingGroupsArray {
		line := *(group.Name)+","
		autoScalingGroups += line

		autoscalingInfo := _get_autoscaling_group_info(nil, *group.Name).AutoScalingGroups[0]
		for _, instance := range autoscalingInfo.Instances {
			instanceTable.Append([]string{*group.Name, *instance.InstanceId, *instance.HealthStatus, *instance.InstanceType, *instance.AvailabilityZone})
		}
	}
	autoScalingGroups = autoScalingGroups[:len(autoScalingGroups)-1]

	// Desired, Min, Max Size ==> int64
	DesiredSize := info.Nodegroup.ScalingConfig.DesiredSize
	MaxSize := info.Nodegroup.ScalingConfig.MaxSize
	MinSize := info.Nodegroup.ScalingConfig.MinSize

	// DiskSize int64
	DiskSize := info.Nodegroup.DiskSize

	table.Append([]string{nodegroup, *Status, instanceTypes, labels, strconv.FormatInt(*MinSize, 10), strconv.FormatInt(*DesiredSize, 10), strconv.FormatInt(*MaxSize, 10), autoScalingGroups, strconv.FormatInt(*DiskSize, 10)})

	table.Render()
	fmt.Println()
	instanceTable.Render()
}

// Get Single Security Group information
func _get_detail_info_of_security_group()  {
	groupId := viper.GetString("securitygroup")
	if len(groupId) <= 0 {
		red("You need to set --securitygroup <group id>")
		os.Exit(1)
	}

	//Get Security Group Information
	info := _get_security_group_detail(nil, []*string{&groupId})

	//Print Security Group
	group := info.SecurityGroups[0]
	table := _print_security_group_info(group)

	table.Render()
}
