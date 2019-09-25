package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type Gateway struct {
	AvailabilityZone string
	NATGWName        string
	NATEIPName       string
	PublicSubnetName string
}

type GuestNATGatewayAdapter struct {
	Gateways  []Gateway
	NATRoutes []NATRoute
}

type NATRoute struct {
	NATGWName             string
	NATRouteName          string
	PrivateRouteTableName string
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	for _, az := range cfg.TenantClusterAvailabilityZones {
		gw := Gateway{
			AvailabilityZone: az.Name,
			NATGWName:        key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATEIPName:       key.SanitizeCFResourceName(key.NATEIPName(az.Name)),
			PublicSubnetName: key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		nr := NATRoute{
			NATGWName:             key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATRouteName:          key.SanitizeCFResourceName(key.NATRouteName(az.Name)),
			PrivateRouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		a.NATRoutes = append(a.NATRoutes, nr)
	}

	return nil
}
