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
	AvailabilityZones []ContextSpecTenantClusterTCCPAvailabilityZone
}

type ContextSpecTenantClusterTCCPAvailabilityZone struct {
	Name       string
	Subnet     ContextSpecTenantClusterTCCPAvailabilityZoneSubnet
	RouteTable ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable
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

type ContextSpecTenantClusterTCCPAvailabilityZoneRouteTable struct {
	Private ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePrivate
	Public  ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic
}

type ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePrivate struct {
	ID string
}

type ContextSpecTenantClusterTCCPAvailabilityZoneRouteTablePublic struct {
	ID string
}

type ContextSpecTenantClusterTCNP struct {
	AvailabilityZones []ContextSpecTenantClusterTCNPAvailabilityZone
}

type ContextSpecTenantClusterTCNPAvailabilityZone struct {
	Name   string
	Subnet ContextSpecTenantClusterTCNPAvailabilityZoneSubnet
}

type ContextSpecTenantClusterTCNPAvailabilityZoneSubnet struct {
	Private ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate
}

type ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate struct {
	CIDR net.IPNet
}
