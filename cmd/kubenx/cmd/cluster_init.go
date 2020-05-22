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
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// clusterInitCmd represents the clusterInit command
var clusterInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initiating the EKS cluster for further usage",
	Long:  `Initiating the EKS cluster for further usage`,
	Run: func(cmd *cobra.Command, args []string) {
		argsLen := len(args)
		if argsLen > 1 {
			Red("Too many Arguments")
			os.Exit(1)
		} else if argsLen == 0 {
			_cluster_initalization("")
		} else {
			cluster_name := args[0]
			_cluster_initalization(cluster_name)
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterInitCmd)
}

// Cluster Initialization
func _cluster_initalization(cluster_name string) {
	// Print the description for initialization
	Red("******* Steps for initialization ********")
	Yellow("Step 1. Tag setup for VPC")
	Yellow("Step 2. Tag setup for public subnet")
	Yellow("Step 3. Tag setup for private subnet")
	Yellow("Step 4. Create Open ID Connector")
	fmt.Println()

	// Get Cluster Information First
	svc := _get_eks_session()
	ec2_svc := _get_ec2_session()
	iam_svc := _get_iam_session()

	// Check the cluster First
	var cluster string
	if len(cluster_name) != 0 {
		cluster = cluster_name
	} else {
		cluster = _choose_cluster()
	}

	// 1. Get Cluster Information
	clusterInfo := _get_cluster_info_with_session(svc, cluster)

	// 2. Get VPC ID and subnets in that VPC
	vpcId := clusterInfo.Cluster.ResourcesVpcConfig.VpcId
	subnetList := _get_subnet_list_in_vpc(ec2_svc, vpcId)

	if len(subnetList.Subnets) == 0 {
		Red("No subnet exists, please checkout out VPC")
		os.Exit(1)
	}

	// 3. Get VPC Information
	vpcInfo := _get_vpc_info(ec2_svc, vpcId)
	tags := vpcInfo.Vpcs[0].Tags

	hasVPCTag := false
	for _, tag := range tags {
		if *tag.Key == "kubernetes.io/cluster/"+cluster && *tag.Value == "shared" {
			hasVPCTag = true
		}
	}

	// Check the vpc tag is updated
	if hasVPCTag {
		Blue("Step 1. VPC Tag is already updated")
	} else {
		Red("Step 1. VPC Tag needs to be updated")
		_update_vpc_tag_for_cluster(ec2_svc, vpcId, cluster)
	}

	// 4. Set Subnet Information

	var publicSubnetIds []*string
	var privateSubnetIds []*string
	for _, subnet := range subnetList.Subnets {
		// Check public subnet
		clusterNameSetup := false
		ELBTypeSetup := false
		isFilteredSubnet := false
		if *subnet.MapPublicIpOnLaunch {
			for _, tag := range subnet.Tags {
				if *tag.Key == "kubernetes.io/cluster/"+cluster && *tag.Value == "shared" {
					clusterNameSetup = true
				}

				if *tag.Key == "kubernetes.io/role/elb" && *tag.Value == "1" {
					ELBTypeSetup = true
				}

				if *tag.Key == "Name" && strings.HasPrefix(*tag.Value, "db") {
					isFilteredSubnet = true
					break
				}
			}

			if !(ELBTypeSetup && clusterNameSetup) && !isFilteredSubnet {
				publicSubnetIds = append(publicSubnetIds, subnet.SubnetId)
			}
		} else {
			for _, tag := range subnet.Tags {
				if *tag.Key == "kubernetes.io/cluster/"+cluster && *tag.Value == "shared" {
					clusterNameSetup = true
				}

				if *tag.Key == "kubernetes.io/role/internal-elb" && *tag.Value == "1" {
					ELBTypeSetup = true
				}

				if *tag.Key == "Name" && strings.HasPrefix(*tag.Value, "db") {
					isFilteredSubnet = true
					break
				}
			}

			if !(ELBTypeSetup && clusterNameSetup) && !isFilteredSubnet {
				privateSubnetIds = append(privateSubnetIds, subnet.SubnetId)
			}
		}

	}

	// Add Tag if there is public subnet which doesn't have the necessary tags
	if len(publicSubnetIds) > 0 {
		Red("Step 2. Tags for Public Subnet needs to be updated")
		_update_subnets_tag_for_cluster(ec2_svc, publicSubnetIds, cluster, "public")
	} else {
		Blue("Step 2. Tags for Public Subnet is already updated")
	}

	// Add Tag if there is private subnet which doesn't have the necessary tags
	if len(privateSubnetIds) > 0 {
		Red("Step 3. Tags for Private Subnet needs to be updated")
		_update_subnets_tag_for_cluster(ec2_svc, privateSubnetIds, cluster, "private")
	} else {
		Blue("Step 3. Tags for Private Subnet is already updated")
	}

	// 5. OpenID Connector Check
	ret, err := _create_openID_connector(iam_svc, clusterInfo.Cluster.Identity.Oidc.Issuer)

	if ret == ALREADY_EXISTS {
		Blue("Step 4. OIDC Provider already exists")
	} else if ret == NEWLY_CREATED {
		Red("Step 4. New OIDC Provider is successfully created")
	} else {
		Red(err)
		os.Exit(1)
	}
}
