package tenantclients

import (
	"context"
)

// EnsureDeleted is not putting the tenant clients into the controller context
// because we do not want to interact with the Tenant Cluster API on delete
// events. This is to reduce eventual friction. Cluster deletion should not be
// affected only because the Tenant Cluster API is not available for some
// reason. Other resources must not rely on tenant clients on delete events.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
