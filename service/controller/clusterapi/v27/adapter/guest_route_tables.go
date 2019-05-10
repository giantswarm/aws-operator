package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

type RouteTableName struct {
	ResourceName        string
	TagName             string
	VPCPeeringRouteName string
}

type GuestRouteTablesAdapter struct {
	HostClusterCIDR        string
	PublicRouteTableName   RouteTableName
	PrivateRouteTableNames []RouteTableName
}

func (r *GuestRouteTablesAdapter) Adapt(cfg Config) error {
	r.HostClusterCIDR = cfg.ControlPlaneVPCCidr
	r.PublicRouteTableName = RouteTableName{
		ResourceName: "PublicRouteTable",
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.MachineDeployment)); i++ {
		rtName := RouteTableName{
			ResourceName:        key.PrivateRouteTableName(i),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: key.VPCPeeringRouteName(i),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
