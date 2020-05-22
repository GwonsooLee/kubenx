package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

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

