package controllercontext

type ContextStatus struct {
	ControlPlane  ContextStatusControlPlane
	TenantCluster ContextStatusTenantCluster
}

type ContextStatusControlPlane struct {
	AWSAccountID string
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
