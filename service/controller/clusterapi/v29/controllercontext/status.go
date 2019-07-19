package controllercontext

import (
	"net"

	"github.com/aws/aws-sdk-go/service/ec2"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type ContextStatus struct {
	ControlPlane  ContextStatusControlPlane
	TenantCluster ContextStatusTenantCluster
}

type ContextStatusControlPlane struct {
	AWSAccountID string
	NATGateway   ContextStatusControlPlaneNATGateway
	RouteTable   ContextStatusControlPlaneRouteTable
	PeerRole     ContextStatusControlPlanePeerRole
	VPC          ContextStatusControlPlaneVPC
}

type ContextStatusControlPlaneNATGateway struct {
	Addresses []*ec2.Address
}

type ContextStatusControlPlaneRouteTable struct {
	// Mappings are key value pairs of control plane route table names and their
	// IDs, where the map keys are route table names and the map values are route
	// table IDs. The mapping is managed by the routetable resource.
	Mappings map[string]string
}

type ContextStatusControlPlanePeerRole struct {
	ARN string
}

type ContextStatusControlPlaneVPC struct {
	CIDR string
}

type ContextStatusTenantCluster struct {
	AWSAccountID          string
	Encryption            ContextStatusTenantClusterEncryption
	HostedZoneNameServers string
	MasterInstance        ContextStatusTenantClusterMasterInstance
	TCCP                  ContextStatusTenantClusterTCCP
	VersionBundleVersion  string
	WorkerInstance        ContextStatusTenantClusterWorkerInstance
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
	ASG               ContextStatusTenantClusterTCCPASG
	AvailabilityZones []ContextStatusTenantClusterTCCPAvailabilityZone
	IsTransitioning   bool
	MachineDeployment v1alpha1.MachineDeployment
	RouteTables       []*ec2.RouteTable
	Subnets           []*ec2.Subnet
	VPC               ContextStatusTenantClusterTCCPVPC
}

type ContextStatusTenantClusterTCCPASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
	Name            string
}

type ContextStatusTenantClusterTCCPAvailabilityZone struct {
	Name   string
	Subnet ContextStatusTenantClusterTCCPAvailabilityZoneSubnet
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

type ContextStatusTenantClusterTCCPVPC struct {
	ID                  string
	PeeringConnectionID string
}

type ContextStatusTenantClusterWorkerInstance struct {
	DockerVolumeSizeGB string
	Image              string
	Type               string
}
