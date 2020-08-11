package runner

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/spf13/viper"
)

// Get EKS Session
func GetEksSession() *eks.EKS {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Get All EKS Cluster
func GetEKSClusterList(svc *eks.EKS) []string {
	if svc == nil {
		svc = GetEksSession()
	}

	inputParams := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	res, err := svc.ListClusters(inputParams)
	if err != nil {
		utils.Red(err)
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
func ChooseCluster() string {

	// Check the cluster First
	var ret string
	ret = viper.GetString("cluster")

	// If cluster is not given then choose!
	if len(ret) <= 0 {
		options := GetEKSClusterList(nil)

		//Exception if there is no cluster in the account
		if len(options) == 0 {
			utils.Yellow("You have no cluster in the account. Please check the account.")
			os.Exit(0)
		}
		prompt := &survey.Select{
			Message: "Choose a cluster:",
			Options: options,
		}
		survey.AskOne(prompt, &ret)

		if ret == "" {
			utils.Red("You canceled the choice")
			os.Exit(1)
		}
	}

	return ret
}

// Get Cluster Information with session
func GetClusterInfoWithSession(svc *eks.EKS, cluster string) *eks.DescribeClusterOutput {
	inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(cluster)}
	ret, err := svc.DescribeCluster(inputParamsDesc)

	if err != nil {
		utils.Red(err)
		os.Exit(1)
	}

	return ret
}

// Get Nodegroup List with eks session
func GetNodeGroupList(svc *eks.EKS, cluster string) []string {
	if svc == nil {
		svc = GetEksSession()
	}

	inputParams := &eks.ListNodegroupsInput{ClusterName: aws.String(cluster), MaxResults: aws.Int64(100)}
	res, err := svc.ListNodegroups(inputParams)
	if err != nil {
		utils.Red(err)
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
func ChooseNodegroup(cluster string) string {
	var ret string
	ret = viper.GetString("nodegroup")

	// If nogegroup is not given then choose!
	if len(ret) <= 0 {
		options := GetNodeGroupList(nil, cluster)

		//Exception if there is no nodegroup in the cluster
		if len(options) == 0 {
			utils.Yellow("You have no nodegroup in the cluster. Please check the cluster.")
			os.Exit(0)
		}

		prompt := &survey.Select{
			Message: "Choose a nodegroup:",
			Options: options,
		}
		survey.AskOne(prompt, &ret)

		if ret == "" {
			utils.Red("You canceled the choice")
			os.Exit(1)
		}
	}

	return ret
}

// Get Cluster Information with session
func GetNodegroupInfoWithSession(svc *eks.EKS, cluster string, nodegroup string) *eks.DescribeNodegroupOutput {
	ngParams := &eks.DescribeNodegroupInput{ClusterName: aws.String(cluster), NodegroupName: aws.String(nodegroup)}
	ret, err := svc.DescribeNodegroup(ngParams)
	if err != nil {
		utils.Red(err)
		os.Exit(1)
	}

	return ret
}
