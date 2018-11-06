package cloudformation

import "github.com/aws/aws-sdk-go/service/cloudformation"

// CloudFormationInterface provides a set of methods to work with
// CloudFormation stacks. *CloudFormation struct from
// "github.com/aws/aws-sdk-go/service/cloudformation" fulfils this interface.
type CloudFormationInterface interface {
	DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
}
