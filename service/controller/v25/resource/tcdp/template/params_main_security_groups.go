package template

type ParamsMainSecurityGroups struct {
	Cluster      ParamsMainSecurityGroupsCluster
	ControlPlane ParamsMainSecurityGroupsControlPlane
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
