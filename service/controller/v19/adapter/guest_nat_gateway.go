package adapter

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/controller/v19/key"
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
	for i := 0; i < key.SpecAvailabilityZones(cfg.CustomObject); i++ {
		gw := Gateway{
			ClusterID:             key.ClusterID(cfg.CustomObject),
			NATGWName:             fmt.Sprintf("NATGateway%02d", i),
			NATEIPName:            fmt.Sprintf("NATEIP%02d", i),
			NATRouteName:          fmt.Sprintf("NATRoute%02d", i),
			PrivateRouteTableName: fmt.Sprintf("PrivateRouteTable%02d", i),
			PublicSubnetName:      fmt.Sprintf("PublicSubnet%02d", i),
		}
		a.Gateways = append(a.Gateways, gw)
	}

	return nil
}
