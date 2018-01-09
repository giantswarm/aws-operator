package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
)

type EC2ClientMock struct {
	unexistingSg         bool
	sgID                 string
	unexistingSubnet     bool
	subnetID             string
	unexistingRouteTable bool
	routeTableID         string
	vpcID                string
	vpcCIDR              string
	unexistingVPC        bool
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
	if e.subnetID == "" {
		e.subnetID = "subnet-1234"
	}

	if e.unexistingSubnet {
		return nil, fmt.Errorf("subnet not found")
	}

	output := &ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			&ec2.Subnet{
				SubnetId: aws.String(e.subnetID),
			},
		},
	}
	return output, nil
}

func (e *EC2ClientMock) DescribeRouteTables(input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	if e.unexistingRouteTable {
		return nil, fmt.Errorf("route table not found")
	}

	output := &ec2.DescribeRouteTablesOutput{
		RouteTables: []*ec2.RouteTable{
			&ec2.RouteTable{
				RouteTableId: aws.String(e.routeTableID),
			},
		},
	}
	return output, nil
}

func (e *EC2ClientMock) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	if e.unexistingVPC {
		return nil, fmt.Errorf("vpc not found")
	}

	output := &ec2.DescribeVpcsOutput{
		Vpcs: []*ec2.Vpc{
			&ec2.Vpc{
				CidrBlock: aws.String(e.vpcCIDR),
				VpcId:     aws.String(e.vpcID),
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
	accountID   string
	isError     bool
	peerRoleArn string
}

func (i *IAMClientMock) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
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
	output := &iam.GetUserOutput{
		User: &iam.User{
			Arn: aws.String("::::" + i.accountID),
		},
	}

	return output, nil
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
			&elb.LoadBalancerDescription{
				DNSName:                   aws.String(e.dns),
				CanonicalHostedZoneNameID: aws.String(e.hostedZone),
			},
		},
	}
	return output, nil
}
