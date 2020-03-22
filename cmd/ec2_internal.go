package cmd
import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/spf13/viper"
	"github.com/olekukonko/tablewriter"
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

// Get Security Group Information
func _get_security_group_detail(svc *ec2.EC2, groupIds []*string) *ec2.DescribeSecurityGroupsOutput{
	if (svc == nil){
		svc = _get_ec2_session()
	}

	inputParam := &ec2.DescribeSecurityGroupsInput{GroupIds: groupIds}
	ret, err := svc.DescribeSecurityGroups(inputParam)

	if err != nil {
		red(err)
		os.Exit(1)
	}

	return ret
}

// Print Security Group information
func _print_security_group_info(group *ec2.SecurityGroup) *tablewriter.Table  {
	var ipProtocol string
	var portRange string

	table := tablewriter.NewWriter(os.Stdout)
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

		if (inbound.IpRanges != nil) {
			for _, cidrObj := range inbound.IpRanges {
				table.Append([]string{typeStr, ipProtocol, portRange, *cidrObj.CidrIp, *cidrObj.Description})
			}
		}

		if (inbound.UserIdGroupPairs != nil) {
			for _, pair := range inbound.UserIdGroupPairs {
				table.Append([]string{typeStr, ipProtocol, portRange, *pair.GroupId, *pair.Description})
			}
		}

	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)
	table.SetCaption(true, *group.GroupName)

	return table
}
