package cloudformation

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type eC2ClientMock struct {
	sgExists bool
	sgID     string
}

func (e *eC2ClientMock) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	if e.sgExists {
		output := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupId: aws.String(e.sgID),
				},
			},
		}
		return output, nil
	}
	return nil, fmt.Errorf("not found")
}

type cFClientMock struct{}

func (c *cFClientMock) CreateStack(*awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error) {
	return nil, nil
}
func (c *cFClientMock) DeleteStack(*awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error) {
	return nil, nil
}
func (c *cFClientMock) DescribeStacks(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
	return nil, nil
}
func (c *cFClientMock) UpdateStack(*awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error) {
	return nil, nil
}

type iAMClientMock struct {
	accountID string
}

func (i *iAMClientMock) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
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
