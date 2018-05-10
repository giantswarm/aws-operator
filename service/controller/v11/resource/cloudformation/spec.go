package cloudformation

import "github.com/aws/aws-sdk-go/service/cloudformation"

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the
	// stack.
	defaultCreationTimeout = 10

	workerRoleKey = "WorkerRole"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name string

	MasterImageID              string
	MasterInstanceType         string
	MasterInstanceResourceName string
	MasterCloudConfigVersion   string
	MasterInstanceMonitoring   bool

	ShouldScale  bool
	ShouldUpdate bool

	Status string

	WorkerCount              string
	WorkerImageID            string
	WorkerInstanceType       string
	WorkerCloudConfigVersion string
	WorkerMonitoring         bool

	UpdateStackInput cloudformation.UpdateStackInput

	VersionBundleVersion string
}
