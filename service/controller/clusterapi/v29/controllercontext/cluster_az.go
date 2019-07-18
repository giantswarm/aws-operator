package controllercontext

import "net"

type ContextTenantClusterAvailabilityZone struct {
	Name          string
	PrivateSubnet ContextTenantClusterAvailabilityZonePrivateSubnet
	PublicSubnet  ContextTenantClusterAvailabilityZonePublicSubnet
}

type ContextTenantClusterAvailabilityZonePublicSubnet struct {
	CIDR net.IPNet
	ID   string
}

type ContextTenantClusterAvailabilityZonePrivateSubnet struct {
	CIDR net.IPNet
	ID   string
}
