package adapter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

type GuestVPCAdapter struct {
	CidrBlock        string
	ClusterID        string
	InstallationName string
	HostAccountID    string
	PeerVPCID        string
	PeerRoleArn      string
	Region           string
	RegionARN        string
	RouteTableNames  []RouteTableName
}

func (v *GuestVPCAdapter) Adapt(cfg Config) error {
	v.CidrBlock = key.StatusNetworkCIDR(cfg.CustomObject)
	v.ClusterID = key.ClusterID(cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.HostAccountID
	v.PeerVPCID = key.PeerID(cfg.CustomObject)
	v.Region = key.Region(cfg.CustomObject)
	v.RegionARN = key.RegionARN(cfg.CustomObject)

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

	PublicRouteTable := RouteTableName{
		ResourceName: key.PublicRouteTableName(0),
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}
	v.RouteTableNames = append(v.RouteTableNames, PublicRouteTable)

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		rtName := RouteTableName{
			ResourceName:        key.PrivateRouteTableName(i),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: key.VPCPeeringRouteName(i),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}
