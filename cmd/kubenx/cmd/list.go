package cmd

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List kubernetes resources",
	Long:  `List kubernetes resources`,
	Run: func(cmd *cobra.Command, args []string) {
		//Get Arguments
		argsLen := len(args)

		if argsLen > 1 {
			Red("Too many Arguments")
			os.Exit(1)
		} else if argsLen == 0 {
			Red("At least one argument is needed.")
			os.Exit(1)
		}

		objType := args[0]

		// Call function according to the second third parameter
		switch {
		case objType == "cluster":
			list_clusters()
		case objType == "nodegroup" || objType == "ng":
			list_nodegroups()
		default:
			Red("Please follow the direction")
		}

	},
}

func init() {
	// Get Flag Values
	listCmd.Flags().StringP("cluster", "c", "", "EKS Cluster Name")

	//Bind Flag to viper
	viper.BindPFlag("cluster", listCmd.Flags().Lookup("cluster"))

}

// List Clusters
func list_clusters() {
	svc := _get_eks_session()
	clusters := _get_eks_cluster_list(svc)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "Version", "Arn", "Endpoint"})
	for _, cluster := range clusters {
		clusterInfo := _get_cluster_info_with_session(svc, cluster)
		arn := clusterInfo.Cluster.Arn
		endpoint := clusterInfo.Cluster.Endpoint
		version := clusterInfo.Cluster.Version
		status := clusterInfo.Cluster.Status

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
		cluster = _choose_cluster()
	}

	svc := _get_eks_session()
	nodegroupList := _get_node_group_list(svc, cluster)

	// Tables for showing outputs
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "STATUS", "INSTANCE TYPE", "LABELS", "MIN SIZE", "DISIRED SIZE", "MAX SIZE", "AUTOSCALING GROUPDS", "DISK SIZE"})
	// Get node group information
	for _, nodegroup := range nodegroupList {
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

		labels := ""
		if len(LabelsObj) > 0 {
			for key, value := range LabelsObj {
				line := key + "=" + *value + ","
				labels += line
			}
			labels = labels[:len(labels)-1]
		}

		// InstanceTypes Array of Map(Object)
		AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
		autoScalingGroups := ""
		for _, group := range AutoScalingGroupsArray {
			line := *(group.Name) + ","
			autoScalingGroups += line
		}
		autoScalingGroups = autoScalingGroups[:len(autoScalingGroups)-1]

		// Desired, Min, Max Size ==> int64
		DesiredSize := info.Nodegroup.ScalingConfig.DesiredSize
		MaxSize := info.Nodegroup.ScalingConfig.MaxSize
		MinSize := info.Nodegroup.ScalingConfig.MinSize

		// DiskSize int64
		DiskSize := info.Nodegroup.DiskSize

		table.Append([]string{nodegroup, *Status, instanceTypes, labels, strconv.FormatInt(*MinSize, 10), strconv.FormatInt(*DesiredSize, 10), strconv.FormatInt(*MaxSize, 10), autoScalingGroups, strconv.FormatInt(*DiskSize, 10)})
	}
	table.Render()
}
