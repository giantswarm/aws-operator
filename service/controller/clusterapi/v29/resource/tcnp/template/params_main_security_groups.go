package template

type ParamsMainSecurityGroups struct {
	ControlPlane  ParamsMainSecurityGroupsControlPlane
	TenantCluster ParamsMainSecurityGroupsTenantCluster
}

type ParamsMainSecurityGroupsControlPlane struct {
	VPC     ParamsMainSecurityGroupsControlPlaneVPC
	Ingress ParamsMainSecurityGroupsControlPlaneIngress
}

type ParamsMainSecurityGroupsControlPlaneVPC struct {
	CIDR string
}

type ParamsMainSecurityGroupsControlPlaneIngress struct {
	ID string
}

type ParamsMainSecurityGroupsTenantCluster struct {
	VPC ParamsMainSecurityGroupsTenantClusterVPC
}

type ParamsMainSecurityGroupsTenantClusterVPC struct {
	ID string
}
