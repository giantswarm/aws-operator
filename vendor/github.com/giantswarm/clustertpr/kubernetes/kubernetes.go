package kubernetes

import (
	"github.com/giantswarm/clustertpr/kubernetes/api"
	"github.com/giantswarm/clustertpr/kubernetes/dns"
	"github.com/giantswarm/clustertpr/kubernetes/hyperkube"
	"github.com/giantswarm/clustertpr/kubernetes/kubelet"
)

type Kubernetes struct {
	API api.API `json:"api" yaml:"api"`
	DNS dns.DNS `json:"dns" yaml:"dns"`
	// Domain is the base domain of the Kubernetes cluster, e.g.
	// g8s.fra-1.giantswarm.io.
	Domain    string              `json:"domain" yaml:"domain"`
	Hyperkube hyperkube.Hyperkube `json:"hyperkube" yaml:"hyperkube"`
	Kubelet   kubelet.Kubelet     `json:"kubelet" yaml:"kubelet"`
}
