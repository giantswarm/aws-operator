package kubernetes

type Kubelet struct {
	// AltNames is the alternative names used to generate certificates for the
	// Kubernetes kubelet. These are usually the alternative names of the API
	// server plus the service name of the API server. The addition is important
	// to make kubelets able to connect to the API servers.
	AltNames string `json:"altNames" yaml:"altNames"`
	// Domain is the API domain for the Kubernetes worker nodes, e.g.
	// worker.<cluster-id>.g8s.fra-1.giantswarm.io.
	Domain string `json:"domain" yaml:"domain"`
	Labels string `json:"labels" yaml:"labels"`
	// Port is the kubelet service port, used in the Kubernetes service definition
	// of the worker nodes.
	Port int `json:"port" yaml:"port"`
}
