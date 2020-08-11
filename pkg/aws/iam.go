package aws

import (
	"encoding/json"
	"fmt"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

type KubenxAussmeConfig struct {
	SessionName      string            `json:"session_name"`
	Assume           map[string]string `json:"assume"`
	EKSAssumeMapping map[string]string `json:"eks-assume-mapping"`
}

var (
	//Constant Value
	ALREADY_EXISTS   = 2
	NEWLY_CREATED    = 1
	CREATION_FAILURE = 0

	//OPEN_ID_CA_FINGERPRINT
	CA_FINGERPRINT   = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
	CONFIG_FILE_PATH = utils.HomeDir() + "/.kubenx/config"
)

func GetIAMSession(role *string) *iam.IAM {
	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())

	var creds *credentials.Credentials
	if role != nil {
		creds = stscreds.NewCredentials(mySession, *role)
	}

	if creds == nil {
		return iam.New(mySession, &aws.Config{Region: aws.String(awsRegion)})
	}
	return iam.New(mySession, &aws.Config{Region: aws.String(awsRegion), Credentials: creds})
}

func getSTSSession() *sts.STS {
	ResetAWSEnvironmentVariable()

	awsRegion := viper.GetString("region")
	mySession := session.Must(session.NewSession())
	svc := sts.New(mySession, &aws.Config{Region: aws.String(awsRegion)})

	return svc
}

func ResetAWSEnvironmentVariable() {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
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

//Find Assume role mapping information
func FindEKSAussmeInfo() (KubenxAussmeConfig, error) {
	kubenxAssumeConfig := KubenxAussmeConfig{}
	rawJson, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		return kubenxAssumeConfig, err
	}

	_ = json.Unmarshal(rawJson, &kubenxAssumeConfig)

	return kubenxAssumeConfig, nil
}

// Create STS Assume Role
func AssumeRole(arn string, session_name string) *sts.Credentials {
	svc := getSTSSession()
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(arn),
		RoleSessionName: aws.String(session_name),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case sts.ErrCodePackedPolicyTooLargeException:
				fmt.Println(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
			case sts.ErrCodeRegionDisabledException:
				fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil
	}

	return result.Credentials
}
