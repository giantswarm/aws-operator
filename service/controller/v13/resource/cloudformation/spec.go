package cloudformation

import "github.com/aws/aws-sdk-go/service/cloudformation"

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the
	// stack.
	defaultCreationTimeout = 10

	workerRoleKey = "WorkerRole"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"

	// masterRoleARNOutputKey is the key of the master role ARN output in the main
	// guest stack.
	masterRoleARNOutputKey = "MasterRoleARN"
	// workerRoleARNOutputKey is the key of the worker role ARN output in the main
	// guest stack.
	workerRoleARNOutputKey = "WorkerRoleARN"
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
	WorkerInstanceMonitoring bool
	WorkerInstanceType       string
	WorkerCloudConfigVersion string

	UpdateStackInput cloudformation.UpdateStackInput

	VersionBundleVersion string
}
