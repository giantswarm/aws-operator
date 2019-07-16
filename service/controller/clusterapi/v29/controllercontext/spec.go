package controllercontext

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	AvailabilityZones []ContextTenantClusterAvailabilityZone
}
