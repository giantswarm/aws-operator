package controllercontext

import "net"

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	TCCP ContextSpecTenantClusterTCCP
	TCNP ContextSpecTenantClusterTCNP
}

type ContextSpecTenantClusterTCCP struct {
	AvailabilityZones []ContextTenantClusterAvailabilityZone
}

type ContextSpecTenantClusterTCNP struct {
	AvailabilityZones []ContextSpecTenantClusterTCNPAvailabilityZone
}

type ContextSpecTenantClusterTCNPAvailabilityZone struct {
	AvailabilityZone string
	PrivateSubnet    net.IPNet
}
