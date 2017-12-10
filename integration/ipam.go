// +build k8srequired

package integration

import (
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
)

const (
	defaultCIDRMask   = 24
	totalBitsLength   = 32
	privateSubnetMask = 25
	publicSubnetMask  = 25
	networkCIDR       = "10.1.0.0/16"
)

type awsVPC struct {
	NetworkCIDR       string
	PrivateSubnetCIDR string
	PublicSubnetCIDR  string
}

func newAWSVPCBlock(client AWSClient) (awsVPC, error) {
	cidrMask := net.CIDRMask(defaultCIDRMask, totalBitsLength)
	existingSubnets, err := listSubnets(client)
	if err != nil {
		return awsVPC{}, microerror.Mask(err)
	}
	cidr, err := newSubnet(cidrMask, existingSubnets)
	if err != nil {
		return awsVPC{}, microerror.Mask(err)
	}

	// Configure private and public subnets as subnets of the new subnet.
	privateSubnetMask := net.CIDRMask(privateSubnetMask, totalBitsLength)
	privateSubnetCIDR, err := ipam.Free(cidr, privateSubnetMask, nil)
	if err != nil {
		return awsVPC{}, microerror.Mask(err)
	}

	publicSubnetMask := net.CIDRMask(publicSubnetMask, totalBitsLength)
	publicSubnetCIDR, err := ipam.Free(cidr, publicSubnetMask, []net.IPNet{privateSubnetCIDR})
	if err != nil {
		return awsVPC{}, microerror.Mask(err)
	}

	vpc := awsVPC{
		NetworkCIDR:       cidr.String(),
		PrivateSubnetCIDR: privateSubnetCIDR.String(),
		PublicSubnetCIDR:  publicSubnetCIDR.String(),
	}
	return vpc, nil
}

func newSubnet(mask net.IPMask, existingSubnets []net.IPNet) (net.IPNet, error) {
	_, network, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	subnet, err := ipam.Free(*network, mask, existingSubnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	return subnet, nil
}

func listSubnets(client AWSClient) ([]net.IPNet, error) {
	output := []net.IPNet{}

	input := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Installation"),
				Values: []*string{
					aws.String("gauss"),
				},
			},
			{
				Name: aws.String("state"),
				Values: []*string{
					aws.String("available"),
				},
			},
		},
	}

	result, err := client.EC2.DescribeVpcs(input)
	if err != nil {
		return output, microerror.Mask(err)
	}

	// this map will help us to have unique entries in the output
	existentCIDR := make(map[string]bool)

	for _, vpc := range result.Vpcs {
		// check if we have already added the current CIDR block
		currentCIDRStr := *vpc.CidrBlock
		if _, ok := existentCIDR[currentCIDRStr]; ok {
			continue
		}
		_, currentCIDR, err := net.ParseCIDR(currentCIDRStr)
		if err != nil {
			return output, microerror.Mask(err)
		}

		existentCIDR[currentCIDRStr] = true
		output = append(output, *currentCIDR)
	}

	return output, nil
}
