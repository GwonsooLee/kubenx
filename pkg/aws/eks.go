package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/spf13/viper"
)

// Get EKS Session
func GetEksSession(role *string) *eks.EKS {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())

	var creds *credentials.Credentials
	if role != nil {
		creds = stscreds.NewCredentials(mySession, *role)
	}

	if creds == nil {
		return eks.New(mySession, &aws.Config{Region: aws.String(awsRegion)})
	}
	return eks.New(mySession, &aws.Config{Region: aws.String(awsRegion), Credentials: creds})
}

// Get Cluster Information with session
func GetClusterInfo(svc *eks.EKS, cluster string) (*eks.DescribeClusterOutput, error) {
	inputParamsDesc := &eks.DescribeClusterInput{Name: aws.String(cluster)}
	ret, err := svc.DescribeCluster(inputParamsDesc)

	if err != nil {
		return nil, err
	}

	return ret, nil
}
