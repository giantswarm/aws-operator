package cloudformationv1

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 10

	workersParameterKey        = "WorkersParam"
	imageIDParameterKey        = "ImageIDParam"
	clusterVersionParameterKey = "ClusterVersionParam"

	workersOutputKey       = "WorkersOutput"
	imageIDOutputKey       = "ImageIDOutput"
	clusterVersionOuputKey = "ClusterVersionOutput"

	cloudFormationTemplatesDirectory = "service/templates/cloudformation"
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name           string
	ImageID        string
	Workers        string
	ClusterVersion string
}
