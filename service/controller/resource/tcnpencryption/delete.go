package tcnpencryption

import (
	"context"
)

// EnsureDeleted is a noop because we do not want to delete anything when Node
// Pools are deleted. Deletion of the Tenant Cluster's KMS key is managed in the
// tccpencryption resource.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
