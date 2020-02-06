package template

type ParamsMainRouteTables struct {
	ClusterID              string
	HostClusterCIDR        string
	PrivateRouteTableNames []ParamsMainRouteTablesRouteTableName
	PublicRouteTableNames  []ParamsMainRouteTablesRouteTableName
}

type ParamsMainRouteTablesRouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
	VPCPeeringRouteName string
}
