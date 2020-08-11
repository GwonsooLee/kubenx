package runner

import (
	"os"

	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/spf13/viper"
)

// Get Autoscaling Session
func getAutoscalingSession() *autoscaling.AutoScaling {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := autoscaling.New(mySession, &aws.Config{Region: aws.String(awsRegion)})
	return svc
}

// Describe Security Group Information
func getAutoscalingGroupInfo(svc *autoscaling.AutoScaling, autoscalingGroupName string) *autoscaling.DescribeAutoScalingGroupsOutput {
	if svc == nil {
		svc = getAutoscalingSession()
	}

	inputParam := &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []*string{aws.String(autoscalingGroupName)}}

	ret, err := svc.DescribeAutoScalingGroups(inputParam)

	if err != nil {
		utils.Red(err)
		os.Exit(1)
	}

	return ret
}
