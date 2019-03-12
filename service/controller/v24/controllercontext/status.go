package controllercontext

type ContextStatus struct {
	ControlPlane  ContextStatusControlPlane
	TenantCluster ContextStatusTenantCluster
}

type ContextStatusControlPlane struct {
	AWSAccountID string
	PeerRole     ContextStatusControlPlanePeerRole
	VPC          ContextStatusControlPlaneVPC
}

type ContextStatusControlPlanePeerRole struct {
	ARN string
}

type ContextStatusControlPlaneVPC struct {
	CIDR string
}

type ContextStatusTenantCluster struct {
	AWSAccountID           string
	TCCP                   ContextStatusTenantClusterTCCP
	EncryptionKey          string
	HostedZoneNameServers  string
	VPCPeeringConnectionID string
}

type ContextStatusTenantClusterTCCP struct {
	ASG ContextStatusTenantClusterTCCPASG
}
