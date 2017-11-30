package cloudformationv1

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 10

	workersParameterKey        = "workers"
	imageIDParameterKey        = "imageID"
	clusterVersionParameterKey = "clusterVersion"

	cloudFormationTemplatesDirectory = "service/templates/cloudformation"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name           string
	ImageID        string
	Workers        string
	ClusterVersion string
}
