package routetable

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2 interface {
	DescribeRouteTables(*ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error)
}
