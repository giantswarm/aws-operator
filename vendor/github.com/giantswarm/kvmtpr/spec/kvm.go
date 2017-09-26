package spec

import "github.com/giantswarm/kvmtpr/spec/kvm"

type KVM struct {
	EndpointUpdater kvm.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          kvm.K8sKVM          `json:"k8sKVM" yaml:"k8sKVM"`
	Masters         []kvm.Node          `json:"masters" yaml:"masters"`
	Workers         []kvm.Node          `json:"workers" yaml:"workers"`
}
