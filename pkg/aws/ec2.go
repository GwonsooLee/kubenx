package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
)

// Get EC2 Session
func GetEC2Session(role *string) *ec2.EC2 {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())

	var creds *credentials.Credentials
	if role != nil {
		creds = stscreds.NewCredentials(mySession, *role)
	}

	if creds == nil {
		return ec2.New(mySession, &aws.Config{Region: aws.String(awsRegion)})
	}
	return ec2.New(mySession, &aws.Config{Region: aws.String(awsRegion), Credentials: creds})
}

// Describe Single VPC Information
func GetVPCInfo(svc *ec2.EC2, vpcId *string) (*ec2.DescribeVpcsOutput, error) {
	var vpcIds []*string
	vpcIds = append(vpcIds, vpcId)
	inputParam := &ec2.DescribeVpcsInput{VpcIds: vpcIds}
	ret, err := svc.DescribeVpcs(inputParam)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Describe subnet subnet Information
func GetSubnetsInfo(svc *ec2.EC2, subnetIds []*string) (*ec2.DescribeSubnetsOutput, error) {
	inputParam := &ec2.DescribeSubnetsInput{SubnetIds: subnetIds}
	ret, err := svc.DescribeSubnets(inputParam)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Get all subnet list in vpc
func GetSubnetListInVPC(svc *ec2.EC2, vpcId *string) (*ec2.DescribeSubnetsOutput, error) {
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
		return nil, err
	}

	return ret, nil
}

// Update VPC Tag for cluster
func UpdateVPCTagForCluster(svc *ec2.EC2, vpcId *string, cluster string) error {
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
		return err
	}

	return nil
}

// Update Subnet Tag for cluster
func UpdateSubnetsTagForCluster(svc *ec2.EC2, subnets []*string, cluster string, subnetType string) error {
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
		return err
	}

	return nil
}
