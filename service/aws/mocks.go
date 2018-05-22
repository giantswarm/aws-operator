package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

const (
	// accountIDIndex represents the index in which we can find the account ID in the user ARN
	// (splitting by colon)
	accountIDIndex  = 4
	accountIDLength = 12
)

type STSClientMock struct {
	stsiface.STSAPI

	accountID string
}

func (i *STSClientMock) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if i.accountID == "" {
		i.accountID = "00"
	}
	// pad accountID to required length
	toPad := accountIDLength - len(i.accountID)
	for j := 0; j < toPad; j++ {
		i.accountID += "0"
	}
	output := &sts.GetCallerIdentityOutput{
		Arn: aws.String("::::" + i.accountID),
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
