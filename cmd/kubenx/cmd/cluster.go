package cmd

import (
	"fmt"
	"os"
	"io"
	"strings"
	"context"
	"github.com/spf13/cobra"
	"github.com/olekukonko/tablewriter"
	"github.com/GwonsooLee/kubenx/pkg/aws"
	"github.com/GwonsooLee/kubenx/pkg/color"
)

var (
	TAG_PREFIX = "kubernetes"
)

//Get new get function
func NewCmdCluster() *cobra.Command {
	return NewCmd("cluster").
		WithDescription("Cluster related command").
		AddCommand(NewCmdInitCluster()).
		SetFlags().
		RunWithNoArgs(execCluster)
}

//Get new get function
func NewCmdGetCluster() *cobra.Command {
	return NewCmd("cluster").
		WithDescription("Get informaton about cluster").
		RunWithNoArgs(execGetCluster)
}

func execCluster(ctx context.Context, out io.Writer) error {
	return nil
}

// Function for getting services
func execGetCluster(ctx context.Context, out io.Writer) error {
	return runExecutorWithAWS(ctx, func(executor Executor) error {
		// Check the cluster First
		cluster, err := GetCurrentCluster()
		if err != nil {
			return err
		}

		// 1. Get Cluster Information
		clusterInfo, err := aws.GetClusterInfo(executor.EKS, cluster)
		if err != nil {
			return err
		}

		clusterTable := tablewriter.NewWriter(os.Stdout)
		clusterTable.SetHeader([]string{"Name", cluster})
		clusterTable.Append([]string{"Version", *clusterInfo.Cluster.Version})
		clusterTable.Append([]string{"Status", *clusterInfo.Cluster.Status})
		clusterTable.Append([]string{"Arn", *clusterInfo.Cluster.Arn})
		clusterTable.Append([]string{"Endpoint", *clusterInfo.Cluster.Endpoint})
		clusterTable.Append([]string{"Cluster SG", *clusterInfo.Cluster.ResourcesVpcConfig.ClusterSecurityGroupId})

		// Get VPC Information
		vpcInfo, err := aws.GetVPCInfo(executor.EC2, clusterInfo.Cluster.ResourcesVpcConfig.VpcId)
		if err != nil {
			return err
		}

		//Find VPC Name with Name Tag
		vpcName := NO_STRING
		for _, obj := range vpcInfo.Vpcs[0].Tags {
			if *obj.Key == "Name" {
				vpcName = *obj.Value
				break
			}
		}
		vpcStr := fmt.Sprintf("%s(%s)", vpcName, *vpcInfo.Vpcs[0].VpcId)
		clusterTable.Append([]string{"VPC ID", vpcStr})
		clusterTable.Append([]string{"VPC Cidr Block", *vpcInfo.Vpcs[0].CidrBlock})

		// Get Subnet List in VPC
		subnetList := clusterInfo.Cluster.ResourcesVpcConfig.SubnetIds
		if len(subnetList) > 0 {
			var subnetIds []*string
			for _, subnetId := range subnetList {
				subnetIds = append(subnetIds, subnetId)
			}

			//Get all subnet information
			subnetInfo, err := aws.GetSubnetsInfo(executor.EC2, subnetIds)
			if err != nil {
				return err
			}

			//Retrieve subnet details
			for _, subnetObj := range subnetInfo.Subnets {
				az := subnetObj.AvailabilityZone
				tags := NO_STRING
				subnetName := NO_STRING
				for _, obj := range subnetObj.Tags {
					if strings.HasPrefix(*obj.Key, TAG_PREFIX) {
						line := fmt.Sprintf("%s=%s ", *obj.Key, *obj.Value)
						tags += line
					} else if *obj.Key == "Name" {
						subnetName = *obj.Value
					}
				}

				clusterTable.Append([]string{subnetName + "(" + *az + ")", tags})
			}
		}

		clusterTable.SetAlignment(tablewriter.ALIGN_LEFT)
		clusterTable.Render()
		return nil
	})
}

//Init Cluster
func NewCmdInitCluster() *cobra.Command {
	return NewCmd("init").
		WithDescription("Initiating the EKS cluster for further usage").
		RunWithNoArgs(execInitCluster)
}

// Function for init cluster services
func execInitCluster(ctx context.Context, out io.Writer) error {
	return runExecutorWithAWS(ctx, func(executor Executor) error {

		// Get Cluster Information First
		// Check the cluster First
		cluster, err := GetCurrentCluster()
		if err != nil {
			return err
		}

		// 1. Get Cluster Information
		clusterInfo, err := aws.GetClusterInfo(executor.EKS, cluster)
		if err != nil {
			return err
		}

		// 2. Get VPC ID and subnets in that VPC
		vpcId := clusterInfo.Cluster.ResourcesVpcConfig.VpcId
		subnetList, err := aws.GetSubnetListInVPC(executor.EC2, vpcId)
		if err != nil {
			return err
		}

		if len(subnetList.Subnets) == 0 {
			return fmt.Errorf("No subnet exists, please checkout out VPC")
		}

		// 3. Get VPC Information
		vpcInfo, err := aws.GetVPCInfo(executor.EC2, vpcId)
		if err != nil {
			return err
		}

		// Tags
		tags := vpcInfo.Vpcs[0].Tags

		hasVPCTag := false
		for _, tag := range tags {
			if *tag.Key == "kubernetes.io/cluster/"+cluster && *tag.Value == "shared" {
				hasVPCTag = true
			}
		}

		// Print the description for initialization
		color.Red.Fprintln(out, "******* Steps for initialization ********")
		color.Yellow.Fprintln(out, "Step 1. Tag setup for VPC")
		color.Yellow.Fprintln(out, "Step 2. Tag setup for public subnet")
		color.Yellow.Fprintln(out, "Step 3. Tag setup for private subnet")
		color.Yellow.Fprintln(out, "Step 4. Create Open ID Connector")
		fmt.Println()


		// Check the vpc tag is updated
		if hasVPCTag {
			color.Blue.Fprintln(out, "Step 1. VPC Tag is already updated")
		} else {
			color.Red.Fprintln(out, "Step 1. VPC Tag needs to be updated")
			aws.UpdateVPCTagForCluster(executor.EC2, vpcId, cluster)
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
			color.Red.Fprintln(out, "Step 2. Tags for Public Subnet needs to be updated")
			aws.UpdateSubnetsTagForCluster(executor.EC2, publicSubnetIds, cluster, "public")
		} else {
			Blue("Step 2. Tags for Public Subnet is already updated")
		}

		// Add Tag if there is private subnet which doesn't have the necessary tags
		if len(privateSubnetIds) > 0 {
			color.Red.Fprintln(out, "Step 3. Tags for Private Subnet needs to be updated")
			aws.UpdateSubnetsTagForCluster(executor.EC2, privateSubnetIds, cluster, "private")
		} else {
			color.Blue.Fprintln(out, "Step 3. Tags for Private Subnet is already updated")
		}

		// 5. OpenID Connector Check
		ret, err := aws.CreateOpenIDConnector(executor.IAM, clusterInfo.Cluster.Identity.Oidc.Issuer)

		if ret == aws.ALREADY_EXISTS {
			color.Blue.Fprintln(out, "Step 4. OIDC Provider already exists")
		} else if ret == aws.NEWLY_CREATED {
			color.Red.Fprintln(out, "Step 4. New OIDC Provider is successfully created")
		} else {
			return err
		}

		return nil
	})
}
