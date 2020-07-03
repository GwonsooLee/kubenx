package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

// Get All EKS Cluster
func getEKSClusterList(svc *eks.EKS) []string {
	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		Red(err)
		os.Exit(1)
	}

	// Change []*string to []string
	var ret []string
	for _, cluster := range res.Clusters {
		ret = append(ret, *cluster)
	}

	return ret
}

// Get EKS Session
func _get_eks_session() *eks.EKS {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Get All EKS Cluster
func _get_eks_cluster_list(svc *eks.EKS) []string {
	if svc == nil {
		svc = _get_eks_session()
	}

	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		Red(err)
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

	// Check the cluster First
	var ret string
	ret = viper.GetString("cluster")

	// If cluster is not given then choose!
	if len(ret) <= 0 {
		options := _get_eks_cluster_list(nil)

		//Exception if there is no cluster in the account
		if len(options) == 0 {
			Yellow("You have no cluster in the account. Please check the account.")
			os.Exit(0)
		}
		prompt := &survey.Select{
			Message: "Choose a cluster:",
			Options: options,
		}
		survey.AskOne(prompt, &ret)

		if ret == "" {
			Red("You canceled the choice")
			os.Exit(1)
		}
	}

	return ret
}

// Get Cluster Information with session
func _get_cluster_info_with_session(svc *eks.EKS, cluster string) *eks.DescribeClusterOutput {
	inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(cluster)}
	ret, err := svc.DescribeCluster(inputParamsDesc)

	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}

// Get Nodegroup List with eks session
func _get_node_group_list(svc *eks.EKS, cluster string) []string {
	if svc == nil {
		svc = _get_eks_session()
	}

	inputParams := &eks.ListNodegroupsInput{ClusterName: aws.String(cluster), MaxResults: aws.Int64(100)}
	res, err := svc.ListNodegroups(inputParams)
	if err != nil {
		Red(err)
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
	ret = viper.GetString("nodegroup")

	// If nogegroup is not given then choose!
	if len(ret) <= 0 {
		options := _get_node_group_list(nil, cluster)

		//Exception if there is no nodegroup in the cluster
		if len(options) == 0 {
			Yellow("You have no nodegroup in the cluster. Please check the cluster.")
			os.Exit(0)
		}

		prompt := &survey.Select{
			Message: "Choose a nodegroup:",
			Options: options,
		}
		survey.AskOne(prompt, &ret)

		if ret == "" {
			Red("You canceled the choice")
			os.Exit(1)
		}
	}

	return ret
}

// Get Cluster Information with session
func _get_nodegroup_info_with_session(svc *eks.EKS, cluster string, nodegroup string) *eks.DescribeNodegroupOutput {
	ngParams := &eks.DescribeNodegroupInput{ClusterName: aws.String(cluster), NodegroupName: aws.String(nodegroup)}
	ret, err := svc.DescribeNodegroup(ngParams)
	if err != nil {
		Red(err)
		os.Exit(1)
	}

	return ret
}
