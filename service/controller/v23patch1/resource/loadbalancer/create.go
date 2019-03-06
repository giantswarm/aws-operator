package loadbalancer

import (
	"context"
)

// EnsureCreated is a no-op, because the loadbalancer resource is only
// interested in delete events.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	return nil
}
