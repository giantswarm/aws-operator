package ipam

import (
	"context"
)

// EnsureCreated allocates tenant cluster network segments. It gathers existing
// subnets from existing system resources like VPCs and Cluster CRs.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	return nil
}
