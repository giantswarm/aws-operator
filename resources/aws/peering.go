package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"
)

const (
	RequesterVpcFilterName = "requester-vpc-info.vpc-id"
	AccepterVpcFilterName  = "accepter-vpc-info.vpc-id"

	accountIDPosition = 4
)

type VPCPeeringConnection struct {
	VPCId     string // VPCId is the ID of the VPC in the guest cluster.
	PeerVPCId string // PeerVPCId the ID of the VPC in the host cluster.
	id        string
	AWSEntity
}

func (v VPCPeeringConnection) findExisting() (*ec2.VpcPeeringConnection, error) {
	vpcPeeringConnections, err := v.Clients.EC2.DescribeVpcPeeringConnections(
		&ec2.DescribeVpcPeeringConnectionsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(RequesterVpcFilterName),
					Values: []*string{
						aws.String(v.VPCId),
					},
				},
				{
					Name: aws.String(AccepterVpcFilterName),
					Values: []*string{
						aws.String(v.PeerVPCId),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(vpcPeeringConnections.VpcPeeringConnections) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, VPCPeeringConnectionType, v.id)
	} else if len(vpcPeeringConnections.VpcPeeringConnections) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return vpcPeeringConnections.VpcPeeringConnections[0], nil
}

func (v *VPCPeeringConnection) checkIfExists() (bool, error) {
	_, err := v.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *VPCPeeringConnection) CreateIfNotExists() (bool, error) {
	exists, err := v.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := v.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *VPCPeeringConnection) CreateOrFail() error {
	resp, err := v.HostClients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	userArn := *resp.User.Arn
	peerOwnerId := strings.Split(userArn, ":")[accountIDPosition]

	vpcPeeringConnection, err := v.Clients.EC2.CreateVpcPeeringConnection(
		&ec2.CreateVpcPeeringConnectionInput{
			PeerOwnerId: &peerOwnerId,
			VpcId:       &v.VPCId,
			PeerVpcId:   &v.PeerVPCId,
		},
	)
	if err != nil {
		return microerror.Mask(err)
	}

	v.id = *vpcPeeringConnection.VpcPeeringConnection.VpcPeeringConnectionId

	if _, err := v.HostClients.EC2.AcceptVpcPeeringConnection(
		&ec2.AcceptVpcPeeringConnectionInput{
			VpcPeeringConnectionId: aws.String(v.id),
		},
	); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (v *VPCPeeringConnection) Delete() error {
	vpcPeeringConnection, err := v.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := v.Clients.EC2.DeleteVpcPeeringConnection(
		&ec2.DeleteVpcPeeringConnectionInput{
			VpcPeeringConnectionId: vpcPeeringConnection.VpcPeeringConnectionId,
		},
	); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (v *VPCPeeringConnection) GetID() (string, error) {
	if v.id != "" {
		return v.id, nil
	}

	vpcPeeringConnection, err := v.findExisting()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *vpcPeeringConnection.VpcPeeringConnectionId, nil
}
