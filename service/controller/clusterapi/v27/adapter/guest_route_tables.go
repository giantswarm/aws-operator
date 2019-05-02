package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
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
		TagName:      legacykey.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}

	for i := 0; i < len(legacykey.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		rtName := RouteTableName{
			ResourceName:        legacykey.PrivateRouteTableName(i),
			TagName:             legacykey.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: legacykey.VPCPeeringRouteName(i),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
