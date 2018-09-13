package adapter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/key"
)

type GuestVPCAdapter struct {
	CidrBlock        string
	ClusterID        string
	InstallationName string
	HostAccountID    string
	PeerVPCID        string
	PeerRoleArn      string
}

func (v *GuestVPCAdapter) Adapt(cfg Config) error {
	v.CidrBlock = cfg.CustomObject.Spec.AWS.VPC.CIDR
	v.ClusterID = clusterID(cfg)
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
	} else if len(output.Vpcs) == 0 {
		return "", microerror.Maskf(notFoundError, "vpc: %s", vpcID)
	} else if len(output.Vpcs) > 1 {
		return "", microerror.Maskf(tooManyResultsError, "vpc: %s found %d vpcs", vpcID, len(output.Vpcs))
	}
	return *output.Vpcs[0].CidrBlock, nil
}
