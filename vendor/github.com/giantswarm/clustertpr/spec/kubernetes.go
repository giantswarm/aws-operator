package spec

import "github.com/giantswarm/clustertpr/spec/kubernetes"

type Kubernetes struct {
	API kubernetes.API `json:"api" yaml:"api"`
	// CloudProvider enables cloud provider specific functionality
	// can be aws, azure, gce, ... needs to be unset for baremetal
	// see https://kubernetes.io/docs/getting-started-guides/scratch/#cloud-providers)
	CloudProvider string         `json:"cloudProvider" yaml:"cloudProvider"`
	DNS           kubernetes.DNS `json:"dns" yaml:"dns"`
	// Domain is the base domain of the Kubernetes cluster, e.g.
	// g8s.fra-1.giantswarm.io.
	Domain            string                       `json:"domain" yaml:"domain"`
	Hyperkube         kubernetes.Hyperkube         `json:"hyperkube" yaml:"hyperkube"`
	IngressController kubernetes.IngressController `json:"ingressController" yaml:"ingressController"`
	Kubectl           kubernetes.Kubectl           `json:"kubectl" yaml:"kubectl"`
	Kubelet           kubernetes.Kubelet           `json:"kubelet" yaml:"kubelet"`
	NetworkSetup      kubernetes.NetworkSetup      `json:"networkSetup" yaml:"networkSetup"`
	SSH               kubernetes.SSH               `json:"ssh" yaml:"ssh"`
}
