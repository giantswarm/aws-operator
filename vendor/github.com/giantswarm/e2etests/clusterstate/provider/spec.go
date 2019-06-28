package provider

import "k8s.io/client-go/kubernetes"

type Clients interface {
	// K8sClient returns a properly configured control plane client for the
	// Kubernetes API.
	K8sClient() kubernetes.Interface
}

type Interface interface {
	RebootMaster() error
	ReplaceMaster() error
}
