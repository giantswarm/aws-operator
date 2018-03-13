package loadbalancer

import (
	"context"
)

// GetDesiredState returns nil as this resource only implements deletion.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
