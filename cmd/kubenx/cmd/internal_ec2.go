package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

// Get EC2 Session
func _get_ec2_session() *ec2.EC2 {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := ec2.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Describe Single VPC Information
func _get_vpc_info(svc *ec2.EC2, vpcId *string) *ec2.DescribeVpcsOutput {
	var vpcIds []*string
	vpcIds = append(vpcIds, vpcId)
	inputParam := &ec2.DescribeVpcsInput{VpcIds: vpcIds}
	ret, err := svc.DescribeVpcs(inputParam)

	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}

// Get all subnet list in vpc
func _get_subnet_list_in_vpc(svc *ec2.EC2, vpcId *string) *ec2.DescribeSubnetsOutput {
	inputParam := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcId},
			},
		},
	}
	ret, err := svc.DescribeSubnets(inputParam)
	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}

// Describe Single subnet Information
func _get_subnet_info(svc *ec2.EC2, subnetIds []*string) *ec2.DescribeSubnetsOutput {
	inputParam := &ec2.DescribeSubnetsInput{SubnetIds: subnetIds}
	ret, err := svc.DescribeSubnets(inputParam)

	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}

// Get Security Group Information
func _get_security_group_detail(svc *ec2.EC2, groupIds []*string) *ec2.DescribeSecurityGroupsOutput {
	if svc == nil {
		svc = _get_ec2_session()
	}

	inputParam := &ec2.DescribeSecurityGroupsInput{GroupIds: groupIds}
	ret, err := svc.DescribeSecurityGroups(inputParam)

	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}

// Print Security Group information
func _print_security_group_info(group *ec2.SecurityGroup) *tablewriter.Table {
	var ipProtocol string
	var portRange string

	table := _get_table_object()
	table.SetHeader([]string{"Type", "Protocol", "Port Range", "Destination", "Description"})

	// Inbound Traffic
	typeStr := "Inbound"
	for _, inbound := range group.IpPermissions {
		// Ip Protocol
		if *inbound.IpProtocol == "-1" {
			ipProtocol = "All Traffic"
			portRange = "All"
		} else {
			ipProtocol = *inbound.IpProtocol
			portRange = strconv.FormatInt(*inbound.FromPort, 10)
		}

		if inbound.IpRanges != nil {
			for _, cidrObj := range inbound.IpRanges {
				table.Append([]string{typeStr, ipProtocol, portRange, *cidrObj.CidrIp, *cidrObj.Description})
			}
		}

		if inbound.UserIdGroupPairs != nil {
			for _, pair := range inbound.UserIdGroupPairs {
				table.Append([]string{typeStr, ipProtocol, portRange, *pair.GroupId, *pair.Description})
			}
		}

	}
	table.SetCaption(true, *group.GroupName)

	return table
}

// Get Detailed Information about nodegroup
func _get_detail_info_of_nodegroup() {
	svc := _get_eks_session()

	// Check the cluster First
	cluster := _get_current_cluster()

	//Choose Nodegroup
	nodegroup := _choose_nodegroup(cluster)

	// NodeGroup Information Table
	table := _get_table_object()
	table.SetHeader([]string{"NAME", "STATUS", "INSTANCE TYPE", "LABELS", "MIN SIZE", "DISIRED SIZE", "MAX SIZE", "AUTOSCALING GROUPDS", "DISK SIZE"})

	// Instance Group Information Table
	instanceTable := _get_table_object()
	instanceTable.SetHeader([]string{"Autoscaling Group", "Instance ID", "Health Status", "Instance Type", "Availability Zone"})

	info := _get_nodegroup_info_with_session(svc, cluster, nodegroup)

	// Retrieve Values
	//Status string
	Status := info.Nodegroup.Status

	// InstanceTypes Array
	InstanceTypesArray := info.Nodegroup.InstanceTypes
	instanceTypes := ""
	for _, iType := range InstanceTypesArray {
		line := *iType + ","
		instanceTypes += line
	}
	instanceTypes = instanceTypes[:len(instanceTypes)-1]

	// Label Map(Object)
	LabelsObj := info.Nodegroup.Labels

	labels := []string{}
	if len(LabelsObj) > 0 {
		for key, value := range LabelsObj {
			labels = append(labels, fmt.Sprintf("%s=%s", key ,*value))
		}
	}

	// InstanceTypes Array of Map(Object)
	AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
	autoScalingGroups := ""
	for _, group := range AutoScalingGroupsArray {
		line := *(group.Name) + ","
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

	table.Append([]string{nodegroup, *Status, instanceTypes, strings.Join(labels,","), strconv.FormatInt(*MinSize, 10), strconv.FormatInt(*DesiredSize, 10), strconv.FormatInt(*MaxSize, 10), autoScalingGroups, strconv.FormatInt(*DiskSize, 10)})

	table.Render()
	fmt.Println()
	instanceTable.Render()
}

// Update VPC Tag for cluster
func _update_vpc_tag_for_cluster(svc *ec2.EC2, vpcId *string, cluster string) {
	inputParam := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(*vpcId),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("kubernetes.io/cluster/" + cluster),
				Value: aws.String("shared"),
			},
		},
	}
	_, err := svc.CreateTags(inputParam)
	if err != nil {
		Red(err)
		os.Exit(1)
	}
}

// Update Subnet Tag for cluster
func _update_subnets_tag_for_cluster(svc *ec2.EC2, subnets []*string, cluster string, subnetType string) {
	var tags []*ec2.Tag
	if subnetType == "public" {
		tags = []*ec2.Tag{
			{
				Key:   aws.String("kubernetes.io/cluster/" + cluster),
				Value: aws.String("shared"),
			},
			{
				Key:   aws.String("kubernetes.io/role/elb"),
				Value: aws.String("1"),
			},
		}
	} else {
		tags = []*ec2.Tag{
			{
				Key:   aws.String("kubernetes.io/cluster/" + cluster),
				Value: aws.String("shared"),
			},
			{
				Key:   aws.String("kubernetes.io/role/internal-elb"),
				Value: aws.String("1"),
			},
		}
	}
	inputParam := &ec2.CreateTagsInput{
		Resources: subnets,
		Tags:      tags,
	}
	_, err := svc.CreateTags(inputParam)
	if err != nil {
		Red(err)
		os.Exit(1)
	}
}

// Get Instance Detailed Information
func _get_ec2_instance_info(instanceIds []*string) *ec2.DescribeInstancesOutput {
	svc := _get_ec2_session()
	inputParam := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}
	ec2List, err := svc.DescribeInstances(inputParam)
	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ec2List
}