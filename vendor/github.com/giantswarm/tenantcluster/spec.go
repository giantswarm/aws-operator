package tenantcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
)

const (
	TillerDefaultNamespace = "giantswarm"
)

type Interface interface {
	// NewG8sClient returns a new generated clientset for a tenant cluster.
	NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error)
	// NewHelmClient returns a new Giant Swarm Helm client for a tenant cluster.
	NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error)
	// NewK8sClient returns a new Kubernetes clientset for a tenant cluster.
	NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error)
}
