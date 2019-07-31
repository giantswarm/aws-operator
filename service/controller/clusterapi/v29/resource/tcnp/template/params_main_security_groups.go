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
	Ingress ParamsMainSecurityGroupsTenantClusterIngress
	Master  ParamsMainSecurityGroupsTenantClusterMaster
	VPC     ParamsMainSecurityGroupsTenantClusterVPC
}

type ParamsMainSecurityGroupsTenantClusterIngress struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterMaster struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterVPC struct {
	ID string
}
