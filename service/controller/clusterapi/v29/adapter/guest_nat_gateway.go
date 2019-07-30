package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type Gateway struct {
	ClusterID             string
	NATGWName             string
	NATEIPName            string
	NATRouteName          string
	PrivateRouteTableName string
	PublicRouteTableName  string
	PublicSubnetName      string
}

type GuestNATGatewayAdapter struct {
	Gateways []Gateway
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	for _, az := range cfg.TenantClusterAvailabilityZones {
		gw := Gateway{
			ClusterID:             key.ClusterID(&cfg.CustomObject),
			NATGWName:             key.SanitizeCFResourceName(key.NATGatewayName(az.Name)),
			NATEIPName:            key.SanitizeCFResourceName(key.NATEIPName(az.Name)),
			NATRouteName:          key.SanitizeCFResourceName(key.NATRouteName(az.Name)),
			PrivateRouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
			PublicRouteTableName:  key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
			PublicSubnetName:      key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
