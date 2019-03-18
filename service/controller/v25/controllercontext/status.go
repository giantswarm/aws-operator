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
}

type ContextStatusTenantCluster struct {
	AWSAccountID           string
	EncryptionKey          string
	HostedZoneNameServers  string
	KMS                    ContextStatusTenantClusterKMS
	MasterInstance         ContextStatusTenantClusterMasterInstance
	TCCP                   ContextStatusTenantClusterTCCP
	VersionBundleVersion   string
	VPC                    ContextStatusTenantClusterVPC
	VPCPeeringConnectionID string
	WorkerInstance         ContextStatusTenantClusterWorkerInstance
}

type ContextStatusTenantClusterKMS struct {
	KeyARN string
}

type ContextStatusTenantClusterMasterInstance struct {
	DockerVolumeResourceName string
	Image                    string
	ResourceName             string
	Type                     string
	CloudConfigVersion       string
}

type ContextStatusTenantClusterTCCP struct {
	ASG         ContextStatusTenantClusterTCCPASG
	RouteTables []*ec2.RouteTable
	Subnets     []*ec2.Subnet
}

type ContextStatusTenantClusterVPC struct {
	ID string
}

type ContextStatusTenantClusterWorkerInstance struct {
	DockerVolumeSizeGB string
	CloudConfigVersion string
	Image              string
	Type               string
}
