package cloudformation

import awsCF "github.com/aws/aws-sdk-go/service/cloudformation"

// StackState is the state representation on which the resource methods work
type StackState struct {
	Name    string
	Outputs []*awsCF.Output
}
