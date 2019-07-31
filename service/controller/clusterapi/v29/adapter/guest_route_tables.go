package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type RouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
	TagName             string
	VPCPeeringRouteName string
}

type GuestRouteTablesAdapter struct {
	HostClusterCIDR        string
	PrivateRouteTableNames []RouteTableName
	PublicRouteTableNames  []RouteTableName
}

func (r *GuestRouteTablesAdapter) Adapt(cfg Config) error {
	r.HostClusterCIDR = cfg.ControlPlaneVPCCidr

	for _, az := range cfg.TenantClusterAvailabilityZones {
		rtName := RouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPublic, az.Name),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		r.PublicRouteTableNames = append(r.PublicRouteTableNames, rtName)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		rtName := RouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, az.Name),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
