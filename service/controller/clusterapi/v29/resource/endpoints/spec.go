package endpoints

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Clients struct {
	EC2 EC2
}

// EC2 describes the methods required to be implemented by a EC2 AWS client.
type EC2 interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}
