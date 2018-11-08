package provider

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

type Interface interface {
	// CreateCluster creates a provider specific tenant cluster CR which the
	// provider specific operator reconciles upon and therefore yields a new
	// tenant cluster. This function does not wait for tenant cluster to be
	// created. The id argument defines the tenant cluster ID.
	CreateCluster(ctx context.Context, id string) error
	// DeleteCluster deletes the provider specific tenant cluster CR identified by
	// the given id argument. The implementation does not wait for the deletion to
	// finish, but just returns after deleting the CR.
	DeleteCluster(ctx context.Context, id string) error
	// GetClusterStatus fetches the current cluster status from the tenant
	// cluster's CR.
	GetClusterStatus(ctx context.Context, id string) (v1alpha1.StatusCluster, error)
	// WaitForClusterCreated waits for the tenant cluster identified by the given
	// ID to be created.
	WaitForClusterCreated(ctx context.Context, id string) error
	// WaitForClusterDeleted waits for the tenant cluster identified by the given
	// ID to be deleted.
	WaitForClusterDeleted(ctx context.Context, id string) error
}
