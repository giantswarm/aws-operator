package template

type Gateway struct {
	AvailabilityZone string
	NATGWName        string
	NATEIPName       string
	PublicSubnetName string
}

type ParamsMainNATGateway struct {
	Gateways  []Gateway
	NATRoutes []NATRoute
}

type NATRoute struct {
	NATGWName             string
	NATRouteName          string
	PrivateRouteTableName string
}
