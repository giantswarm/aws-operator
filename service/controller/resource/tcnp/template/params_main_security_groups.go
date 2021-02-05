package template

type ParamsMainSecurityGroups struct {
	ClusterID     string
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
	Ingress     ParamsMainSecurityGroupsTenantClusterIngress
	InternalAPI ParamsMainSecurityGroupsTenantClusterInternalAPI
	Master      ParamsMainSecurityGroupsTenantClusterMaster
	NodePools   []ParamsMainSecurityGroupsTenantClusterNodePool
	VPC         ParamsMainSecurityGroupsTenantClusterVPC
	AWSCNI      ParamsMainSecurityGroupsTenantClusterAWSCNI
}

type ParamsMainSecurityGroupsTenantClusterIngress struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterInternalAPI struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterMaster struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterAWSCNI struct {
	ID string
}

type ParamsMainSecurityGroupsTenantClusterNodePool struct {
	ID           string
	ResourceName string
}

type ParamsMainSecurityGroupsTenantClusterVPC struct {
	ID   string
	CIDR string
}
