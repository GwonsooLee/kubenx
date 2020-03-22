package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/spf13/viper"
	"github.com/AlecAivazis/survey/v2"
)


// Get EKS Session
func _get_eks_session() *eks.EKS {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}


// Get All EKS Cluster
func _get_eks_cluster_list(svc *eks.EKS) []string {
	if(svc == nil){
		svc = _get_eks_session()
	}

	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	// Change []*string to []string
	var ret []string
	for _, cluster := range res.Clusters {
		ret = append(ret, *cluster)
	}

	return ret
}

// Choose cluster
func _choose_cluster() string {
	var ret string
	options := _get_eks_cluster_list(nil)
	prompt := &survey.Select{
		Message: "Choose a cluster:",
		Options: options,
	}
	survey.AskOne(prompt, &ret)

	if (ret == "") {
		red("You canceled the choice")
		os.Exit(1)
	}

	return ret
}

// Get Cluster Information with session
func _get_cluster_info_with_session(svc *eks.EKS, cluster string) *eks.DescribeClusterOutput {
	inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(cluster)}
	ret, err := svc.DescribeCluster(inputParamsDesc)

	if err != nil {
		red(err)
		os.Exit(1)
	}

	return ret
}

// Get Nodegroup List with eks session
func _get_node_group_list(svc *eks.EKS, cluster string) []string {
	if(svc == nil){
		svc = _get_eks_session()
	}

	inputParams := &eks.ListNodegroupsInput{ClusterName: aws.String(cluster), MaxResults: aws.Int64(100)}
	res, err := svc.ListNodegroups(inputParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	// Change []*string to []string
	var ret []string
	for _, ng := range res.Nodegroups {
		ret = append(ret, *ng)
	}

	return ret
}

// Choose cluster
func _choose_nodegroup(cluster string) string {
	var ret string
	options := _get_node_group_list(nil, cluster)
	prompt := &survey.Select{
		Message: "Choose a nodegroup:",
		Options: options,
	}
	survey.AskOne(prompt, &ret)

	if (ret == "") {
		red("You canceled the choice")
		os.Exit(1)
	}

	return ret
}

// Get Cluster Information with session
func _get_nodegroup_info_with_session(svc *eks.EKS, cluster string,nodegroup string) *eks.DescribeNodegroupOutput {
	ngParams := &eks.DescribeNodegroupInput{ClusterName: aws.String(cluster), NodegroupName: aws.String(nodegroup)}
	ret, err := svc.DescribeNodegroup(ngParams)
	if err != nil {
		red(err)
		os.Exit(1)
	}

	return ret
}
