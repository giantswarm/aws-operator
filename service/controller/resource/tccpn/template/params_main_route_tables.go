package template

type ParamsMainRouteTables struct {
	ClusterID              string
	HostClusterCIDR        string
	PrivateRouteTableNames []ParamsMainRouteTablesRouteTableName
	PublicRouteTableNames  []ParamsMainRouteTablesRouteTableName
	VPCID                  string
	PeeringConnectionID    string
}

type ParamsMainRouteTablesRouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
	VPCPeeringRouteName string
}
