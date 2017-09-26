package kvmtpr

import (
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/kvmtpr/spec"
)

type Spec struct {
	Cluster clustertpr.Spec `json:"cluster" yaml:"cluster"`
	KVM     spec.KVM        `json:"kvm" yaml:"kvm"`
}
