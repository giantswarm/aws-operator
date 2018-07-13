package v1alpha1

const (
	DrainerConfigStatusStatusTrue  = "True"
	DrainerConfigStatusTypeDrained = "Drained"
)

func (s DrainerConfigStatus) HasFinalCondition() bool {
	for _, c := range s.Conditions {
		if c.Type == DrainerConfigStatusTypeDrained && c.Status == DrainerConfigStatusStatusTrue {
			return true
		}
	}

	return false
}

func (s DrainerConfigStatus) NewFinalCondition() DrainerConfigStatusCondition {
	return DrainerConfigStatusCondition{
		Status: DrainerConfigStatusStatusTrue,
		Type:   DrainerConfigStatusTypeDrained,
	}
}
