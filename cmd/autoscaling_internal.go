package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/spf13/viper"
)

// Get Autoscaling Session
func _get_autoscaling_session() *autoscaling.AutoScaling {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := autoscaling.New(mySession, &aws.Config{Region: aws.String(awsRegion)})
	return svc
}

// Describe Security Group Information
func _get_autoscaling_group_info(svc *autoscaling.AutoScaling, autoscalingGroupName string) *autoscaling.DescribeAutoScalingGroupsOutput {
	if (svc == nil){
		svc = _get_autoscaling_session()
	}

	inputParam := &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []*string{aws.String(autoscalingGroupName)}}

	ret, err := svc.DescribeAutoScalingGroups(inputParam)

	if err != nil {
		red(err)
		os.Exit(1)
	}

	return ret
}
