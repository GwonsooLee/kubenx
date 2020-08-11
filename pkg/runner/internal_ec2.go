package runner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GwonsooLee/kubenx/pkg/utils"
)

// Get Detailed Information about nodegroup
func GetDetailInfoOfNodegroup() {
	svc := GetEksSession()

	// Check the cluster First
	cluster := getCurrentCluster()

	//Choose Nodegroup
	nodegroup := ChooseNodegroup(cluster)

	// NodeGroup Information Table
	table := utils.GetTableObject()
	table.SetHeader([]string{"NAME", "STATUS", "INSTANCE TYPE", "LABELS", "MIN SIZE", "DISIRED SIZE", "MAX SIZE", "AUTOSCALING GROUPDS", "DISK SIZE"})

	// Instance Group Information Table
	instanceTable := utils.GetTableObject()
	instanceTable.SetHeader([]string{"Autoscaling Group", "Instance ID", "Health Status", "Instance Type", "Availability Zone"})

	info := GetNodegroupInfoWithSession(svc, cluster, nodegroup)

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
			labels = append(labels, fmt.Sprintf("%s=%s", key, *value))
		}
	}

	// InstanceTypes Array of Map(Object)
	AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
	autoScalingGroups := ""
	for _, group := range AutoScalingGroupsArray {
		line := *(group.Name) + ","
		autoScalingGroups += line

		autoscalingInfo := getAutoscalingGroupInfo(nil, *group.Name).AutoScalingGroups[0]
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

	table.Append([]string{nodegroup, *Status, instanceTypes, strings.Join(labels, ","), strconv.FormatInt(*MinSize, 10), strconv.FormatInt(*DesiredSize, 10), strconv.FormatInt(*MaxSize, 10), autoScalingGroups, strconv.FormatInt(*DiskSize, 10)})

	table.Render()
	fmt.Println()
	instanceTable.Render()
}
