package controllercontext

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
)

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	Release v1alpha1.Release
	TCCP    ContextSpecTenantClusterTCCP
	TCNP    ContextSpecTenantClusterTCNP
}

type ContextSpecTenantClusterTCCP struct {
	AvailabilityZones []ContextSpecTenantClusterTCCPAvailabilityZone
}

type ContextSpecTenantClusterTCCPAvailabilityZone struct {
	Name   string
	Subnet ContextSpecTenantClusterTCCPAvailabilityZoneSubnet
}

type ContextSpecTenantClusterTCCPAvailabilityZoneSubnet struct {
	AWSCNI  ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI
	Private ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPrivate
	Public  ContextSpecTenantClusterTCCPAvailabilityZoneSubnetPublic
}

type ContextSpecTenantClusterTCCPAvailabilityZoneSubnetAWSCNI struct {
	CIDR net.IPNet
	ID   string
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
