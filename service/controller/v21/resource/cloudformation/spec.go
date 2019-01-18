package cloudformation

import "github.com/aws/aws-sdk-go/service/cloudformation"

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the
	// stack.
	defaultCreationTimeout = 20

	workerRoleKey = "WorkerRole"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"

	// versionBundleVersionParameterKey is the key name of the Cloud Formation
	// parameter that sets the version bundle version.
	versionBundleVersionParameterKey = "VersionBundleVersionParameter"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name string

	HostedZoneNameServers string

	DockerVolumeResourceName   string
	MasterImageID              string
	MasterInstanceType         string
	MasterInstanceResourceName string
	MasterCloudConfigVersion   string
	MasterInstanceMonitoring   bool

	ShouldScale  bool
	ShouldUpdate bool

	WorkerCount              string
	WorkerDockerVolumeSizeGB int
	WorkerImageID            string
	WorkerInstanceMonitoring bool
	WorkerInstanceType       string
	WorkerCloudConfigVersion string

	UpdateStackInput cloudformation.UpdateStackInput

	VersionBundleVersion string
}
