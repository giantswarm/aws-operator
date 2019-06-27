package provider

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

type Clients interface {
	// G8sClient returns a properly configured control plane client for the Giant
	// Swarm API Extensions Types.
	G8sClient() versioned.Interface
}

type Interface interface {
	CurrentStatus() (v1alpha1.StatusCluster, error)
	CurrentVersion() (string, error)
	NextVersion() (string, error)
	UpdateVersion(nextVersion string) error
}
