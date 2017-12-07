package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type EC2ClientMock struct {
	unexistingSg     bool
	sgID             string
	unexistingSubnet bool
	clusterID        string
}

func (e *EC2ClientMock) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	if !e.unexistingSg {
		output := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupId: aws.String(e.sgID),
				},
			},
		}
		return output, nil
	}
	return nil, fmt.Errorf("security group not found")
}

func (e *EC2ClientMock) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	if e.clusterID == "" {
		e.clusterID = "test-cluster"
	}

	if !e.unexistingSubnet {
		tagNameValue := *input.Filters[0].Values[0]
		if tagNameValue != e.clusterID+"-public" {
			return nil, fmt.Errorf("unexpected tag name value %v", tagNameValue)
		}

		output := &ec2.DescribeSubnetsOutput{
			Subnets: []*ec2.Subnet{
				&ec2.Subnet{
					SubnetId: aws.String("subnet-1234"),
				},
			},
		}
		return output, nil
	}
	return nil, fmt.Errorf("subnet not found")
}

type CFClientMock struct{}

func (c *CFClientMock) CreateStack(*awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error) {
	return nil, nil
}
func (c *CFClientMock) DeleteStack(*awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error) {
	return nil, nil
}
func (c *CFClientMock) DescribeStacks(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
	return nil, nil
}
func (c *CFClientMock) UpdateStack(*awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error) {
	return nil, nil
}

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
