package v1alpha1

const (
	NodeConfigStatusStatusTrue  = "True"
	NodeConfigStatusTypeDrained = "Drained"
)

func (s NodeConfigStatus) HasFinalConditiion() bool {
	for _, c := range s.Conditions {
		if c.Type == NodeConfigStatusTypeDrained && c.Status == NodeConfigStatusStatusTrue {
			return true
		}
	}

	return false
}

func (s NodeConfigStatus) NewFinalConditiion() NodeConfigStatusCondition {
	return NodeConfigStatusCondition{
		Status: NodeConfigStatusStatusTrue,
		Type:   NodeConfigStatusTypeDrained,
	}
}
