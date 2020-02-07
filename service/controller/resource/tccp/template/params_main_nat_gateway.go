package template

type ParamsMainNATGateway struct {
	Gateways  []ParamsMainNATGatewayGateway
	NATRoutes []ParamsMainNATGatewayNATRoute
}

type ParamsMainNATGatewayGateway struct {
	AvailabilityZone string
	ClusterID        string
	NATGWName        string
	NATEIPName       string
	PublicSubnetName string
}

type ParamsMainNATGatewayNATRoute struct {
	NATGWName             string
	NATRouteName          string
	PrivateRouteTableName string
}
