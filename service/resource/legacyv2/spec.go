package legacyv2

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 10

	workersOutputKey        = "WorkersOutput"
	imageIDOutputKey        = "ImageIDOutput"
	clusterVersionOutputKey = "ClusterVersionOutput"
	workerRoleKey           = "workerRole"

	cloudFormationTemplatesDirectory = "service/templates/cloudformation"

	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name           string
	ImageID        string
	Workers        string
	ClusterVersion string
}
