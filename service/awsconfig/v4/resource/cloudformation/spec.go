package cloudformation

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the
	// stack.
	defaultCreationTimeout = 10

	masterImageIDOutputKey            = "MasterImageID"
	masterInstanceTypeOutputKey       = "MasterInstanceType"
	masterCloudConfigVersionOutputKey = "MasterCloudConfigVersion"
	workersOutputKey                  = "WorkerCount"
	workerImageIDOutputKey            = "WorkerImageID"
	workerInstanceTypeOutputKey       = "WorkerInstanceType"
	workerCloudConfigVersionOutputKey = "WorkerCloudConfigVersion"

	workerRoleKey = "WorkerRole"

	cloudFormationGuestTemplatesDirectory    = "service/templates/cloudformation/guest"
	cloudFormationHostPreTemplatesDirectory  = "service/templates/cloudformation/host-pre"
	cloudFormationHostPostTemplatesDirectory = "service/templates/cloudformation/host-post"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name string

	MasterImageID            string
	MasterInstanceType       string
	MasterCloudConfigVersion string

	WorkerCount              string
	WorkerImageID            string
	WorkerInstanceType       string
	WorkerCloudConfigVersion string
}
