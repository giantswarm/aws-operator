package template

type ParamsMainSubnets []ParamsMainSubnetsSubnet

type ParamsMainSubnetsSubnet struct {
	AvailabilityZone string
	CIDR             string
	Name             string
	TenantCluster    ParamsMainSubnetsSubnetTenantCluster
}

type ParamsMainSubnetsSubnetTenantCluster struct {
	VPC ParamsMainSubnetsSubnetTenantClusterVPC
}

type ParamsMainSubnetsSubnetTenantClusterVPC struct {
	ID string
}
