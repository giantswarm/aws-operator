package kubernetes

type Kubernetes struct {
	API API `json:"api" yaml:"api"`
	DNS DNS `json:"dns" yaml:"dns"`
	// Domain is the base domain of the Kubernetes cluster, e.g.
	// g8s.fra-1.giantswarm.io.
	Domain    string    `json:"domain" yaml:"domain"`
	Hyperkube Hyperkube `json:"hyperkube" yaml:"hyperkube"`
	Kubelet   Kubelet   `json:"kubelet" yaml:"kubelet"`
}
