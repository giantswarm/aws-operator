package template

type ParamsMainRouteTables struct {
	ClusterID       string
	HostClusterCIDR string

	AWSCNIRouteTableNames  []ParamsMainRouteTablesRouteTableName
	PrivateRouteTableNames []ParamsMainRouteTablesRouteTableName
	PublicRouteTableNames  []ParamsMainRouteTablesRouteTableName
}

type ParamsMainRouteTablesRouteTableName struct {
	AvailabilityZone       string
	AvailabilityZoneRegion string
	ResourceName           string
	VPCPeeringRouteName    string
}
