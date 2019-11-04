package controllercontext

import (
	"net"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type ContextStatus struct {
	ControlPlane  ContextStatusControlPlane
	TenantCluster ContextStatusTenantCluster
}

type ContextStatusControlPlane struct {
	AWSAccountID string
	NATGateway   ContextStatusControlPlaneNATGateway
	RouteTables  []*ec2.RouteTable
	PeerRole     ContextStatusControlPlanePeerRole
	VPC          ContextStatusControlPlaneVPC
}

type ContextStatusControlPlaneNATGateway struct {
	Addresses []*ec2.Address
}

type ContextStatusControlPlanePeerRole struct {
	ARN string
}

type ContextStatusControlPlaneVPC struct {
	CIDR string
	ID   string
}

type ContextStatusTenantCluster struct {
	AWS                   ContextStatusTenantClusterAWS
	Encryption            ContextStatusTenantClusterEncryption
	HostedZoneNameServers string
	MasterInstance        ContextStatusTenantClusterMasterInstance
	TCCP                  ContextStatusTenantClusterTCCP
	TCNP                  ContextStatusTenantClusterTCNP
	OperatorVersion       string
}

type ContextStatusTenantClusterAWS struct {
	AccountID string
	Region    string
}

type ContextStatusTenantClusterEncryption struct {
	Key string
}

type ContextStatusTenantClusterMasterInstance struct {
	DockerVolumeResourceName string
	Image                    string
	ResourceName             string
	Type                     string
}

type ContextStatusTenantClusterTCCP struct {
	AvailabilityZones []ContextStatusTenantClusterTCCPAvailabilityZone
	IsTransitioning   bool
	NATGateways       []*ec2.NatGateway
	RouteTables       []*ec2.RouteTable
	SecurityGroups    []*ec2.SecurityGroup
	Subnets           []*ec2.Subnet
	VPC               ContextStatusTenantClusterTCCPVPC
}

type ContextStatusTenantClusterTCCPAvailabilityZone struct {
	Name       string
	Subnet     ContextStatusTenantClusterTCCPAvailabilityZoneSubnet
	RouteTable ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable
}

type ContextStatusTenantClusterTCCPAvailabilityZoneSubnet struct {
	Private ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate
	Public  ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic
}

type ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate struct {
	CIDR net.IPNet
	ID   string
}

type ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic struct {
	CIDR net.IPNet
	ID   string
}

type ContextStatusTenantClusterTCCPAvailabilityZoneRouteTable struct {
	Public ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic
}

type ContextStatusTenantClusterTCCPAvailabilityZoneRouteTablePublic struct {
	ID string
}

type ContextStatusTenantClusterTCCPVPC struct {
	ID                  string
	PeeringConnectionID string
}

type ContextStatusTenantClusterTCNP struct {
	ASG            ContextStatusTenantClusterTCNPASG
	WorkerInstance ContextStatusTenantClusterTCNPWorkerInstance
}

type ContextStatusTenantClusterTCNPASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
	Name            string
}

type ContextStatusTenantClusterTCNPWorkerInstance struct {
	DockerVolumeSizeGB string
	Image              string
	Type               string
}
