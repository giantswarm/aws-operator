package cloudformation

// StackState is the state representation on which the resource methods work
type StackState struct {
	Name         string
	TemplateBody string
}
