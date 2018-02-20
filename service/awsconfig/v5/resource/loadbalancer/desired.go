package loadbalancer

import (
	"context"
)

// GetDesiredState returns an empty state as this resource only implements
// deletion.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	emptyState := &LoadBalancerState{
		LoadBalancerNames: []string{},
	}

	return emptyState, nil
}
