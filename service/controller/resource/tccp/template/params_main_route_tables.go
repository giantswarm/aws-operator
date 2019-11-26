package template

type RouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
	VPCPeeringRouteName string
}

type ParamsMainRouteTables struct {
	HostClusterCIDR        string
	PrivateRouteTableNames []RouteTableName
	PublicRouteTableNames  []RouteTableName
}
