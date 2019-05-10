package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	v.CidrBlock = key.StatusClusterNetworkCIDR(cfg.CustomObject)
	v.ClusterID = key.ClusterID(cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.ControlPlaneAccountID
	v.PeerVPCID = cfg.ControlPlaneVPCID
	v.Region = key.Region(cfg.CustomObject)
	v.RegionARN = key.RegionARN(cfg.CustomObject)
	v.PeerRoleArn = cfg.ControlPlanePeerRoleARN

	PublicRouteTable := RouteTableName{
		ResourceName: key.PublicRouteTableName(0),
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}
	v.RouteTableNames = append(v.RouteTableNames, PublicRouteTable)

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.MachineDeployment)); i++ {
		rtName := RouteTableName{
			ResourceName:        key.PrivateRouteTableName(i),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: key.VPCPeeringRouteName(i),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}
