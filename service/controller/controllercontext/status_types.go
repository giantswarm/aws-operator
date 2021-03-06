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
	ASG             ContextStatusTenantClusterASG
	AWS             ContextStatusTenantClusterAWS
	DNS             ContextStatusTenantClusterDNS
	MasterInstance  ContextStatusTenantClusterMasterInstance
	S3Object        ContextStatusTenantClusterS3Object
	TCCP            ContextStatusTenantClusterTCCP
	TCCPN           ContextStatusTenantClusterTCCPN
	TCNP            ContextStatusTenantClusterTCNP
	OperatorVersion string
	ReleaseVersion  string
}

type ContextStatusTenantClusterASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
	Name            string
}

type ContextStatusTenantClusterAWS struct {
	AccountID string
	Region    string
}

type ContextStatusTenantClusterDNS struct {
	APIPublicLoadBalancer     string
	HostedZoneID              string
	HostedZoneNameServers     string
	IngressPublicLoadBalancer string
	InternalHostedZoneID      string
}

type ContextStatusTenantClusterMasterInstance struct {
	EtcdVolumeSnapshotID string
	Type                 string
}

type ContextStatusTenantClusterS3Object struct {
	Uploaded bool
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
	Name   string
	Subnet ContextStatusTenantClusterTCCPAvailabilityZoneSubnet
}

type ContextStatusTenantClusterTCCPAvailabilityZoneSubnet struct {
	AWSCNI  ContextStatusTenantClusterTCCPAvailabilityZoneSubnetAWSCNI
	Private ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPrivate
	Public  ContextStatusTenantClusterTCCPAvailabilityZoneSubnetPublic
}

type ContextStatusTenantClusterTCCPAvailabilityZoneSubnetAWSCNI struct {
	CIDR net.IPNet
	ID   string
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

type ContextStatusTenantClusterTCCPN struct {
	IsTransitioning bool
	InstanceType    string
	MasterReplicas  int
}

type ContextStatusTenantClusterTCNP struct {
	Instances        ContextStatusTenantClusterTCNPInstances
	SecurityGroupIDs []string
	WorkerInstance   ContextStatusTenantClusterTCNPWorkerInstance
}

type ContextStatusTenantClusterTCNPInstances struct {
	InstanceTypes         []string
	NumberOfSpotInstances int
}

type ContextStatusTenantClusterTCNPWorkerInstance struct {
	DockerVolumeSizeGB string
	Image              string
	Type               string
}
