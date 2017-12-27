package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

func VPCID(clients Clients, name string) (string, error) {
	describeVpcInput := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeVpcs(describeVpcInput)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(output.Vpcs) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}

	return *output.Vpcs[0].VpcId, nil
}

func SubnetID(clients Clients, name string) (string, error) {
	describeSubnetInput := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSubnets(describeSubnetInput)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(output.Subnets) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}

	return *output.Subnets[0].SubnetId, nil
}

func RouteTableID(clients Clients, name string) (string, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeRouteTables(input)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(output.RouteTables) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}

	return *output.RouteTables[0].RouteTableId, nil
}
