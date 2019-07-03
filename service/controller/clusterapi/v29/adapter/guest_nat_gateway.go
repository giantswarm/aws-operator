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
	for _, az := range key.WorkerAvailabilityZones(cfg.MachineDeployment) {
		gw := Gateway{
			ClusterID:             key.ClusterID(cfg.CustomObject),
			NATGWName:             key.NATGatewayName(az),
			NATEIPName:            key.NATEIPName(az),
			NATRouteName:          key.NATRouteName(az),
			PrivateRouteTableName: key.PrivateRouteTableName(az),
			PublicSubnetName:      key.PublicSubnetName(az),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
