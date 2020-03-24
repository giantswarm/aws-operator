package v1alpha1

import (
	"time"
)

// DeepCopyTime implements the deep copy logic for time.Time which the k8s
// codegen is not able to generate out of the box.
type DeepCopyTime struct {
	time.Time
}

// DeepCopyDuration implements the deep copy logic for time.Duration which the k8s
// codegen is not able to generate out of the box.
type DeepCopyDuration struct {
	time.Time
}

func (in *DeepCopyDuration) DeepCopyInto(out *DeepCopyDuration) {
	*out = *in
}

func (in *DeepCopyTime) DeepCopyInto(out *DeepCopyTime) {
	*out = *in
}
