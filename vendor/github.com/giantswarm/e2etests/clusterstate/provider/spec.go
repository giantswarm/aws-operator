package provider

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type Clients interface {
	// K8sClient returns a properly configured control plane client for the
	// Kubernetes API.
	K8sClient() kubernetes.Interface
}

type Interface interface {
	RebootMaster() error
	ReplaceMaster() error
	GetClusterAZs(ctx context.Context) ([]string, error)
	ExpectedAZs() ([]string, error)
}
