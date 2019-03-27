package adapter

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

// APIWhitelist defines guest cluster k8s api whitelisting.
type APIWhitelist struct {
	Enabled    bool
	SubnetList string
}

type Hydrater func(config Config) error

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
	// TODO the cloud config versions shouldn't be injected here. These should
	// actually always only be the ones the operator has hard coded. No other
	// version should be used here ever.
	MasterCloudConfigVersion string
	MasterInstanceMonitoring bool

	// TODO the cloud config versions shouldn't be injected here. These should
	// actually always only be the ones the operator has hard coded. No other
	// version should be used here ever.
	WorkerCloudConfigVersion  string
	WorkerDesired             int
	WorkerDockerVolumeSizeGB  string
	WorkerKubeletVolumeSizeGB string
	WorkerLogVolumeSizeGB     string
	WorkerImageID             string
	WorkerInstanceMonitoring  bool
	WorkerInstanceType        string
	WorkerMax                 int
	WorkerMin                 int

	VersionBundleVersion string
}

// SmallCloudconfigConfig represents the data structure required for executing
// the small cloudconfig template.
type SmallCloudconfigConfig struct {
	InstanceRole string
	S3URL        string
}
