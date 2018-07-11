package ipam

import "context"

// EnsureDeleted is a NOP for IPAM resource as allocated subnet will get
// released when the guest cluster VPC and AWSConfig CR is deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
