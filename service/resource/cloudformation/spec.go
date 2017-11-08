package cloudformation

import awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"

const (
	// defaultCreationTimeout is the timeout in minutes for the creation of the stack.
	defaultCreationTimeout = 30
)

// StackState is the state representation on which the resource methods work.
type StackState struct {
	Name    string
	Outputs []*awscloudformation.Output
}
