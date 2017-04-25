package ingress

type IngressController struct {
	// Domain is the external domain of the Ingress Controller running in the
	// Kubernetes cluster, e.g. <cluster-id>.fra-1.gigantic.io.
	Domain string `json:"domain" yaml:"domain"`
	// InsecurePort is the HTTP node port of the Ingress Controller.
	InsecurePort int `json:"insecurePort" yaml:"insecurePort"`
	// SecurePort is the HTTPS node port of the Ingress Controller.
	SecurePort int `json:"securePort" yaml:"securePort"`
}
