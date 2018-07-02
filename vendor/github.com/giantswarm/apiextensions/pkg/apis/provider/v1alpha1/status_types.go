package v1alpha1

import "time"

type StatusCluster struct {
	Conditions []StatusClusterCondition `json:"conditions" yaml:"conditions"`
	Versions   []StatusClusterVersion   `json:"versions" yaml:"versions"`
}

// StatusClusterCondition expresses the conditions in which a guest cluster may
// is.
type StatusClusterCondition struct {
	// Status may be True, False or Unknown.
	Status string `json:"status" yaml:"status"`
	// Type may be Creating, Created, Scaling, Scaled, Draining, Drained,
	// Deleting, Deleted.
	Type string `json:"type" yaml:"type"`
}

// StatusClusterVersion expresses the versions in which a guest cluster was and
// is.
type StatusClusterVersion struct {
	// Date is the time of the given guest cluster version being updated.
	Date time.Time `json:"date" yaml:"date"`
	// Semver is some semver version, e.g. 1.0.0.
	Semver string `json:"semver" yaml:"semver"`
}

// DeepCopyInto implements the deep copy magic the k8s codegen is not able to
// generate out of the box.
func (in *StatusClusterVersion) DeepCopyInto(out *StatusClusterVersion) {
	*out = *in
}
