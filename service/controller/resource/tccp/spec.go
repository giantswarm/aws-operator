package tccp

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type templateParams struct {
	DockerVolumeResourceName   string
	MasterInstanceResourceName string
}

const (
	// asgMaxBatchSizeRatio is the % of instances to be updated during a
	// rolling update.
	asgMaxBatchSizeRatio = 0.3
	// asgMinInstancesRatio is the % of instances to keep in service during a
	// rolling update.
	asgMinInstancesRatio = 0.7
	// defaultEBSVolumeMountPoint is the path for mounting the EBS volume.
	defaultEBSVolumeMountPoint = "/dev/xvdh"
	// defaultEBSVolumeSize is expressed in GB.
	defaultEBSVolumeSize = "100"
	// defaultEBSVolumeType is the EBS volume type.
	defaultEBSVolumeType = "gp2"
	// rollingUpdatePauseTime is how long to pause ASG operations after creating
	// new instances. This allows time for new nodes to join the cluster.
	rollingUpdatePauseTime = "PT15M"
	// logEBSVolumeMountPoint is the path for mounting the log EBS volume
	logEBSVolumeMountPoint = "/dev/xvdf"
	// kubeletEBSVolumeMountPoint is the path for mounting the log EBS volume
	kubeletEBSVolumeMountPoint = "/dev/xvdg"

	// Subnet keys
	subnetDescription = "description"
	subnetGroupName   = "group-name"

	// accountIDIndex represents the index in which we can find the account ID in the user ARN
	// (splitting by colon)
	accountIDIndex = 4

	// The number of seconds AWS will wait, before issuing a health check on
	// instances in an Auto Scaling Group.
	gracePeriodSeconds = 10

	tagKeyName = "Name"

	suffixPublic  = "public"
	suffixPrivate = "private"

	externalELBScheme = "internet-facing"
	internalELBScheme = "internal"

	httpPort  = 80
	httpsPort = 443
)

// APIWhitelist defines guest cluster k8s public/private api whitelisting.
type APIWhitelist struct {
	Private Whitelist
	Public  Whitelist
}

// TODO we copy this because of a circular import issue with the cloudformation
// resource. The way how the resource works with the adapter and how infromation
// is passed has to be reworked at some point. Just hacking this now to keep
// going and to keep the changes as minimal as possible.
type StackState struct {
	Name string

	DockerVolumeResourceName   string
	MasterImageID              string
	MasterInstanceType         string
	MasterInstanceResourceName string
	MasterInstanceMonitoring   bool

	OperatorVersion string
}

// Whitelist represents the structure required for defining whitelisting for
// resource security group
type Whitelist struct {
	Enabled    bool
	SubnetList string
}

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 2
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2
)

const (
	allPorts             = -1
	cadvisorPort         = 4194
	etcdPort             = 2379
	kubeletPort          = 10250
	nodeExporterPort     = 10300
	kubeStateMetricsPort = 10301
	sshPort              = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"
)

type securityConfig struct {
	APIWhitelist                    APIWhitelist
	ControlPlaneNATGatewayAddresses []*ec2.Address
	ControlPlaneVPCCidr             string
	CustomObject                    v1alpha1.Cluster
}
