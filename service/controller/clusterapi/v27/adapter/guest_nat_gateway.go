package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
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
	for i := 0; i < len(legacykey.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		gw := Gateway{
			ClusterID:             legacykey.ClusterID(cfg.CustomObject),
			NATGWName:             legacykey.NATGatewayName(i),
			NATEIPName:            legacykey.NATEIPName(i),
			NATRouteName:          legacykey.NATRouteName(i),
			PrivateRouteTableName: legacykey.PrivateRouteTableName(i),
			PublicSubnetName:      legacykey.PublicSubnetName(i),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
