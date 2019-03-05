package cpf

import "github.com/aws/aws-sdk-go/service/cloudformation"

type CloudFormation interface {
	CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	UpdateTerminationProtection(*cloudformation.UpdateTerminationProtectionInput) (*cloudformation.UpdateTerminationProtectionOutput, error)
	WaitUntilStackCreateComplete(*cloudformation.DescribeStacksInput) error
}
