package kubernetes

type Kubernetes struct {
	API API `json:"api"`
	DNS DNS `json:"dns"`
	// Domain is the base domain of the Kubernetes cluster, e.g.
	// g8s.fra-1.giantswarm.io.
	Domain    string    `json:"domain"`
	Hyperkube Hyperkube `json:"hyperkube"`
	Kubelet   Kubelet   `json:"kubelet"`
}
