package stackoutput

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2 interface {
	DescribeVpcPeeringConnections(*ec2.DescribeVpcPeeringConnectionsInput) (*ec2.DescribeVpcPeeringConnectionsOutput, error)
}
