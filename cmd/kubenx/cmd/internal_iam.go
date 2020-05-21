package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/spf13/viper"
)

// Get EC2 Session
func _get_iam_session() *iam.IAM {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := iam.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Create Open ID Connector
func _create_openID_connector(svc *iam.IAM, issuerUrl *string) (int, error) {
	if svc == nil {
		svc = _get_iam_session()
	}

	inputParam := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
		ThumbprintList: []*string{aws.String(CA_FINGERPRINT)},
		Url:            issuerUrl,
	}
	_, err := svc.CreateOpenIDConnectProvider(inputParam)
	if err != nil {
		awsErr, _ := err.(awserr.Error)
		if awsErr.Code() == "EntityAlreadyExists" {
			return ALREADY_EXISTS, nil
		} else {
			return CREATION_FAILURE, awsErr
		}
	}

	return NEWLY_CREATED, nil
}
