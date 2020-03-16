package controllercontext

import (
	"net"
)

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	TCCP ContextSpecTenantClusterTCCP
	TCNP ContextSpecTenantClusterTCNP
}

type ContextSpecTenantClusterTCCP struct {
	AvailabilityZones []ContextSpecTenantClusterTCCPAvailabilityZone
}

type ContextSpecTenantClusterTCCPAvailabilityZone struct {
	Name   string
	Subnet ContextSpecTenantClusterTCCPAvailabilityZoneSubnet
}

type ContextSpecTenantClusterTCCPAvailabilityZoneSubnet struct {
	Private ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate
	Public  ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic
}

type ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate struct {
	CIDR net.IPNet
	ID   string
}

type ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic struct {
	CIDR net.IPNet
	ID   string
}

type ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePrivate struct {
	ID string
}

type ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic struct {
	ID string
}

type ContextSpecTenantClusterTCNP struct {
	AvailabilityZones []ContextSpecTenantClusterTCNPAvailabilityZone
	SecurityGroupIDs  []string
}

type ContextSpecTenantClusterTCNPAvailabilityZone struct {
	Name       string
	NATGateway ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway
	Subnet     ContextSpecTenantClusterTCNPAvailabilityZoneSubnet
}

type ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway struct {
	ID string
}

type ContextSpecTenantClusterTCNPAvailabilityZoneSubnet struct {
	Private ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate
}

type ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate struct {
	CIDR net.IPNet
}
