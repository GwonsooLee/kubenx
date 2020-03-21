package cmd
import (
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"

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
		red(err)
		os.Exit(1)
	}

	return ret
}

// Describe Single subnet Information
func _get_subnet_info(svc *ec2.EC2, subnetIds []*string) *ec2.DescribeSubnetsOutput {
	inputParam := &ec2.DescribeSubnetsInput{SubnetIds: subnetIds}
	ret, err := svc.DescribeSubnets(inputParam)

	if err != nil {
		red(err)
		os.Exit(1)
	}

	return ret
}
