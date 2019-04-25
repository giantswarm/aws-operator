package drainfinisher

import (
	"context"
)

// EnsureDeleted is a no-op, because the lifecycle resource only has to act on
// create and update events in order to drain guest cluster nodes.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
