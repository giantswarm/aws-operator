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
	PublicSubnetName      string
}

type GuestNATGatewayAdapter struct {
	Gateways []Gateway
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	for _, az := range cfg.TenantClusterAvailabilityZones {
		gw := Gateway{
			ClusterID:             key.ClusterID(&cfg.CustomObject),
			NATGWName:             key.SanitizeCFResourceName(key.NATGatewayName(az)),
			NATEIPName:            key.SanitizeCFResourceName(key.NATEIPName(az)),
			NATRouteName:          key.SanitizeCFResourceName(key.NATRouteName(az)),
			PrivateRouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az)),
			PublicSubnetName:      key.SanitizeCFResourceName(key.PublicSubnetName(az)),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
