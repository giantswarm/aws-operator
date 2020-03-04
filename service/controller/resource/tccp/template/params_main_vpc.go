package template

type ParamsMainVPC struct {
	CidrBlock        string
	ClusterID        string
	InstallationName string
	HostAccountID    string
	PeerVPCID        string
	PeerRoleArn      string
	Region           string
	RegionARN        string
	RouteTableNames  []ParamsMainVPCRouteTableName
}

type ParamsMainVPCRouteTableName struct {
	AvailabilityZone    string
	ResourceName        string
	VPCPeeringRouteName string
}
