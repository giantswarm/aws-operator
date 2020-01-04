package tenantcluster

import (
	"context"

	"k8s.io/client-go/rest"
)

type Interface interface {
	// NewRestConfig returns a Kubernetes REST config for the specified tenant
	// cluster.
	NewRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error)
}
