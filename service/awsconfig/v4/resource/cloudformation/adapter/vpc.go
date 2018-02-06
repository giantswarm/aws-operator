package adapter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v4/key"
)

// template related to this adapter: service/templates/cloudformation/guest/vpc.yaml

type vpcAdapter struct {
	CidrBlock        string
	InstallationName string
	HostAccountID    string
	PeerVPCID        string
	PeerRoleArn      string
}

func (v *vpcAdapter) getVpc(cfg Config) error {
	v.CidrBlock = cfg.CustomObject.Spec.AWS.VPC.CIDR
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.HostAccountID
	v.PeerVPCID = key.PeerID(cfg.CustomObject)

	// PeerRoleArn.
	roleName := key.PeerAccessRoleName(cfg.CustomObject)
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	output, err := cfg.HostClients.IAM.GetRole(input)
	if err != nil {
		return microerror.Mask(err)
	}
	v.PeerRoleArn = *output.Role.Arn

	return nil
}

func VpcCIDR(clients Clients, vpcID string) (string, error) {
	describeVpcInput := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcID),
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
	return *output.Vpcs[0].CidrBlock, nil
}
