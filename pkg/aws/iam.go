package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/viper"
)

var (
	//Constant Value
	ALREADY_EXISTS   = 2
	NEWLY_CREATED    = 1
	CREATION_FAILURE = 0

	//OPEN_ID_CA_FINGERPRINT
	CA_FINGERPRINT = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

)

func GetIAMSession() *iam.IAM {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := iam.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

// Create Open ID Connector
func CreateOpenIDConnector(svc *iam.IAM, issuerUrl *string) (int, error) {
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
