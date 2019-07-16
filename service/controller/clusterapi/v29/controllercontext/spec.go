package controllercontext

import "net"

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	AvailabilityZones []ContextSpecTenantClusterAvailabilityZone
}

type ContextSpecTenantClusterAvailabilityZone struct {
	Name          string
	PrivateSubnet net.IPNet
	PublicSubnet  net.IPNet
}
