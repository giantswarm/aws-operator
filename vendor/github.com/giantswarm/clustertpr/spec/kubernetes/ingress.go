package kubernetes

import "github.com/giantswarm/clustertpr/spec/kubernetes/ingress"

type IngressController struct {
	// Docker is the docker image for the Ingress Controller.
	Docker ingress.Docker `json:"docker" yaml:"docker"`
	// Domain is the external domain of the Ingress Controller running in the
	// Kubernetes cluster, e.g. <cluster-id>.fra-1.gigantic.io.
	Domain string `json:"domain" yaml:"domain"`
	// Wildcard domain is the external domain that is a CNAME to the ingress
	// domain in Kubernetes cluster, e.g. *.<cluster-id>.fra-1.gigantic.io.
	// It will allow to use ingress without creating DNS records.
	WildcardDomain string `json:"wildcardDomain" yaml:"wildcardDomain"`
	// InsecurePort is the HTTP node port of the Ingress Controller.
	InsecurePort int `json:"insecurePort" yaml:"insecurePort"`
	// SecurePort is the HTTPS node port of the Ingress Controller.
	SecurePort int `json:"securePort" yaml:"securePort"`
}
