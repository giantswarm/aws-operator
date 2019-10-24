package template

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type Gateway struct {
	AvailabilityZone string
	NATGWName        string
	NATEIPName       string
	PublicSubnetName string
}

type ParamsNATGateway struct {
	Gateways  []Gateway
	NATRoutes []NATRoute
}

type NATRoute struct {
	NATGWName             string
	NATRouteName          string
	PrivateRouteTableName string
}