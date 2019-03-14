package template

type ParamsMainSecurityGroups struct {
	Cluster       ParamsMainSecurityGroupsCluster
	ControlPlane  ParamsMainSecurityGroupsControlPlane
	TenantCluster ParamsMainSecurityGroupsTenantCluster
}

type ParamsMainSecurityGroupsCluster struct {
	ID string
}

type ParamsMainSecurityGroupsControlPlane struct {
	VPC ParamsMainSecurityGroupsControlPlaneVPC
}

type ParamsMainSecurityGroupsControlPlaneVPC struct {
	CIDR string
}

type ParamsMainSecurityGroupsTenantCluster struct {
	VPC ParamsMainSecurityGroupsTenantClusterVPC
}

type ParamsMainSecurityGroupsTenantClusterVPC struct {
	ID string
}
