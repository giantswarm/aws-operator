package template

type ParamsMainSecurityGroups struct {
	ControlPlane  ParamsMainSecurityGroupsControlPlane
	TenantCluster ParamsMainSecurityGroupsTenantCluster
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
