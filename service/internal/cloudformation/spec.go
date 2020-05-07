package cloudformation

import "github.com/aws/aws-sdk-go/service/cloudformation"

// CF provides a set of methods to work with CloudFormation stacks.
// *CloudFormation struct from
// "github.com/aws/aws-sdk-go/service/cloudformation" fulfils this interface.
type CF interface {
	DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
}
