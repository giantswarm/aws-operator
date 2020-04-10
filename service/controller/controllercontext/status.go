package controllercontext

import "github.com/aws/aws-sdk-go/service/ec2"

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
	ID   string
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
	IgnitionHash             string
	Image                    string
	ResourceName             string
	Type                     string
}

type ContextStatusTenantClusterTCCP struct {
	ASG             ContextStatusTenantClusterTCCPASG
	IsTransitioning bool
	RouteTables     []*ec2.RouteTable
	Subnets         []*ec2.Subnet
	VPC             ContextStatusTenantClusterTCCPVPC
}

type ContextStatusTenantClusterTCCPVPC struct {
	ID                  string
	PeeringConnectionID string
}

type ContextStatusTenantClusterWorkerInstance struct {
	DockerVolumeSizeGB string
	IgnitionHash       string
	Image              string
	Type               string
}
