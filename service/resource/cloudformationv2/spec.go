package cloudformationv2

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 10

	masterImageIDOutputKey            = "MasterImageID"
	masterCloudConfigVersionOutputKey = "MasterCloudConfigVersion"
	workersOutputKey                  = "Workers"
	workerImageIDOutputKey            = "WorkerImageID"
	workerCloudConfigVersionOutputKey = "WorkerCloudConfigVersion"

	workerRoleKey = "WorkerRole"

	cloudFormationGuestTemplatesDirectory    = "service/templates/cloudformation/guest"
	cloudFormationHostPreTemplatesDirectory  = "service/templates/cloudformation/host-pre"
	cloudFormationHostPostTemplatesDirectory = "service/templates/cloudformation/host-post"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name                     string
	MasterImageID            string
	MasterCloudConfigVersion string
	Workers                  string
	WorkerImageID            string
	WorkerCloudConfigVersion string
}
