package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

type RouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
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
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		r.PublicRouteTableNames = append(r.PublicRouteTableNames, rtName)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		rtName := RouteTableName{
			AvailabilityZone:    az.Name,
			ResourceName:        key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			VPCPeeringRouteName: key.SanitizeCFResourceName(key.VPCPeeringRouteName(az.Name)),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
