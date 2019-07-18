package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
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
		ResourceName: key.SanitizeCFResourceName(key.PublicRouteTableName(key.MasterAvailabilityZone(cfg.CustomObject))),
		TagName:      key.PublicRouteTableName(key.MasterAvailabilityZone(cfg.CustomObject)),
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		rtName := RouteTableName{
			ResourceName:        key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, az.Name),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
