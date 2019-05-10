package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
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
	v.CidrBlock = legacykey.StatusNetworkCIDR(cfg.CustomObject)
	v.ClusterID = legacykey.ClusterID(cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.ControlPlaneAccountID
	v.PeerVPCID = cfg.ControlPlaneVPCID
	v.Region = legacykey.Region(cfg.CustomObject)
	v.RegionARN = legacykey.RegionARN(cfg.CustomObject)
	v.PeerRoleArn = cfg.ControlPlanePeerRoleARN

	PublicRouteTable := RouteTableName{
		ResourceName: legacykey.PublicRouteTableName(0),
		TagName:      legacykey.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}
	v.RouteTableNames = append(v.RouteTableNames, PublicRouteTable)

	for i := 0; i < len(legacykey.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		rtName := RouteTableName{
			ResourceName:        legacykey.PrivateRouteTableName(i),
			TagName:             legacykey.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: legacykey.VPCPeeringRouteName(i),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}
