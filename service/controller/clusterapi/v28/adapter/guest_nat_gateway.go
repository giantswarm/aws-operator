package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	for i := 0; i < len(key.StatusAvailabilityZones(cfg.MachineDeployment)); i++ {
		gw := Gateway{
			ClusterID:             key.ClusterID(cfg.CustomObject),
			NATGWName:             key.NATGatewayName(i),
			NATEIPName:            key.NATEIPName(i),
			NATRouteName:          key.NATRouteName(i),
			PrivateRouteTableName: key.PrivateRouteTableName(i),
			PublicSubnetName:      key.PublicSubnetName(i),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
