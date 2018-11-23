package adapter

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/controller/v20/key"
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
	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ.
	a.Gateways = []Gateway{
		{
			ClusterID:             key.ClusterID(cfg.CustomObject),
			NATGWName:             "NATGateway",
			NATEIPName:            "NATEIP",
			NATRouteName:          "NATRoute",
			PrivateRouteTableName: "PrivateRouteTable",
			PublicSubnetName:      "PublicSubnet",
		},
	}

	for i := 1; i < key.SpecAvailabilityZones(cfg.CustomObject); i++ {
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
