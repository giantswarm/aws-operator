package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type EC2ClientMock struct {
	ec2iface.EC2API

	unexistingSg        bool
	sgID                string
	unexistingSubnet    bool
	subnetID            string
	matchingRouteTables int
	routeTableID        string
	vpcID               string
	vpcCIDR             string
	unexistingVPC       bool
	peeringID           string
	elasticIPs          []string
}

func (e *EC2ClientMock) DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	addresses := make([]*ec2.Address, 0)

	for _, eip := range e.elasticIPs {
		address := &ec2.Address{
			PublicIp: aws.String(eip),
		}

		addresses = append(addresses, address)
	}

	output := &ec2.DescribeAddressesOutput{
		Addresses: addresses,
	}

	return output, nil
}

func (e *EC2ClientMock) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	if !e.unexistingSg {
		output := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				{
					GroupId: aws.String(e.sgID),
				},
			},
		}
		return output, nil
	}

	return nil, fmt.Errorf("security group not found")
}

func (e *EC2ClientMock) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	if e.subnetID == "" {
		e.subnetID = "subnet-1234"
	}

	if e.unexistingSubnet {
		return nil, fmt.Errorf("subnet not found")
	}

	output := &ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			{
				SubnetId: aws.String(e.subnetID),
			},
		},
	}
	return output, nil
}

func (e *EC2ClientMock) SetMatchingRouteTables(value int) {
	e.matchingRouteTables = value
}

func (e *EC2ClientMock) DescribeRouteTables(input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	if e.matchingRouteTables == 0 {
		return nil, fmt.Errorf("route table not found")
	}

	rts := []*ec2.RouteTable{}

	for i := 0; i < e.matchingRouteTables; i++ {
		rt := &ec2.RouteTable{
			RouteTableId: aws.String(fmt.Sprintf("%s_%d", e.routeTableID, i)),
		}
		rts = append(rts, rt)
	}

	output := &ec2.DescribeRouteTablesOutput{
		RouteTables: rts,
	}
	return output, nil
}

func (e *EC2ClientMock) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	if e.unexistingVPC {
		return nil, fmt.Errorf("vpc not found")
	}

	output := &ec2.DescribeVpcsOutput{
		Vpcs: []*ec2.Vpc{
			{
				CidrBlock: aws.String(e.vpcCIDR),
				VpcId:     aws.String(e.vpcID),
			},
		},
	}

	return output, nil
}

func (e *EC2ClientMock) DescribeVpcPeeringConnections(*ec2.DescribeVpcPeeringConnectionsInput) (*ec2.DescribeVpcPeeringConnectionsOutput, error) {
	output := &ec2.DescribeVpcPeeringConnectionsOutput{
		VpcPeeringConnections: []*ec2.VpcPeeringConnection{
			{
				VpcPeeringConnectionId: aws.String(e.peeringID),
			},
		},
	}
	return output, nil
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
	iamiface.IAMAPI

	isError     bool
	peerRoleArn string
}

func (i *IAMClientMock) GetRole(input *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	if i.isError {
		return nil, fmt.Errorf("error")
	}
	output := &iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String(i.peerRoleArn),
		},
	}

	return output, nil
}

type KMSClientMock struct {
	kmsiface.KMSAPI

	keyARN  string
	isError bool
}

func (k *KMSClientMock) DescribeKey(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	if k.isError {
		return nil, fmt.Errorf("error")
	}

	output := &kms.DescribeKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			Arn: aws.String(k.keyARN),
		},
	}
	return output, nil
}

type ELBClientMock struct {
	elbiface.ELBAPI

	dns        string
	hostedZone string
	name       string
	isError    bool
}

func (e *ELBClientMock) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	if e.isError {
		return nil, fmt.Errorf("error")
	}
	output := &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				DNSName:                   aws.String(e.dns),
				CanonicalHostedZoneNameID: aws.String(e.hostedZone),
			},
		},
	}
	return output, nil
}

type CloudFormationMock struct{}

func (c *CloudFormationMock) CreateStack(*awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error) {
	return nil, nil
}
func (c *CloudFormationMock) DeleteStack(*awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error) {
	return nil, nil
}
func (c *CloudFormationMock) DescribeStacks(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
	return nil, nil
}
func (c *CloudFormationMock) UpdateStack(*awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error) {
	return nil, nil
}
func (c *CloudFormationMock) UpdateTerminationProtection(*awscloudformation.UpdateTerminationProtectionInput) (*awscloudformation.UpdateTerminationProtectionOutput, error) {
	return nil, nil
}
func (c *CloudFormationMock) WaitUntilStackCreateComplete(*awscloudformation.DescribeStacksInput) error {
	return nil
}
func (c *CloudFormationMock) WaitUntilStackCreateCompleteWithContext(ctx aws.Context, input *awscloudformation.DescribeStacksInput, opts ...request.WaiterOption) error {
	return nil
}

type STSClientMock struct {
	stsiface.STSAPI

	accountID string
	isError   bool
}

func (i *STSClientMock) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if i.isError {
		return nil, fmt.Errorf("error")
	}
	if i.accountID == "" {
		i.accountID = "00"
	}
	// pad accountID to required length
	toPad := accountIDLength - len(i.accountID)
	for j := 0; j < toPad; j++ {
		i.accountID += "0"
	}
	output := &sts.GetCallerIdentityOutput{
		Arn: aws.String("::::" + i.accountID),
	}

	return output, nil
}
