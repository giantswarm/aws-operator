package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	// accountIDIndex represents the index in which we can find the account ID in the user ARN
	// (splitting by colon)
	accountIDIndex  = 4
	accountIDLength = 12
)

type IAMClientMock struct {
	accountID string
}

func (i *IAMClientMock) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
	if i.accountID == "" {
		i.accountID = "00"
	}
	// pad accountID to required length
	toPad := accountIDLength - len(i.accountID)
	for j := 0; j < toPad; j++ {
		i.accountID += "0"
	}
	output := &iam.GetUserOutput{
		User: &iam.User{
			Arn: aws.String("::::" + i.accountID),
		},
	}

	return output, nil
}

type AwsServiceMock struct {
	AccountID string
	KeyArn    string
	IsError   bool
}

func (a AwsServiceMock) GetAccountID() (string, error) {
	if a.IsError {
		return "", fmt.Errorf("error!!")
	}

	return a.AccountID, nil
}

func (a AwsServiceMock) GetKeyArn(clusterID string) (string, error) {
	if a.IsError {
		return "", fmt.Errorf("error!!")
	}

	return a.KeyArn, nil
}
