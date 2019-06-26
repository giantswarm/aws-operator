package v1alpha1

const (
	ClusterVersionLimit = 5
)

const (
	ClusterStatusConditionCreated  = "Created"
	ClusterStatusConditionCreating = "Creating"
)

const (
	ClusterStatusConditionDeleted  = "Deleted"
	ClusterStatusConditionDeleting = "Deleting"
)

const (
	ClusterStatusConditionUpdated  = "Updated"
	ClusterStatusConditionUpdating = "Updating"
)

// CommonClusterStatus is shared type to contain provider independent cluster status
// information.
type CommonClusterStatus struct {
	Conditions []CommonClusterStatusCondition `json:"conditions" yaml:"conditions"`
	ID         string                         `json:"id" yaml:"id"`
	Versions   []CommonClusterStatusVersion   `json:"versions" yaml:"versions"`
}

type CommonClusterStatusCondition struct {
	LastTransitionTime DeepCopyTime `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Condition          string       `json:"condition" yaml:"condition"`
}

type CommonClusterStatusVersion struct {
	LastTransitionTime DeepCopyTime `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Version            string       `json:"version" yaml:"version"`
}
