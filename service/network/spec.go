package network

import (
	"context"
	"net"
)

// AllocationCallbacks holds function pointers for two methods that must be
// called inside a lock when allocating new network range.
type AllocationCallbacks struct {
	// GetReservedNetworks implementation must return all networks that are
	// allocated on any given moment. Failing to do that will result in
	// overlapping allocations.
	GetReservedNetworks func(context.Context) ([]net.IPNet, error)

	// PersistAllocatedNetwork must mutate shared persistent state so that on
	// successful execution persistet network is visible when
	// GetReservedNetworks() is called.
	PersistAllocatedNetwork func(context.Context, net.IPNet) error
}

// Allocator is an interface for IPAM implementation that manages network range
// and subnets for an installation.
type Allocator interface {
	// Allocate performs the subnet allocation from given fullRange for a CIDR
	// block with netSize. It requires function pointers to callbacks that are
	// used for getting currently reserved networks and persisting allocated
	// network. These are called within an single lock so given that callbacks
	// work correctly, this is concurrent safe within single process e.g. when
	// logic between multiple controllers are allocating networks from the same
	// pool.
	//
	// NOTE: This is NOT concurrent safe between distinct processes.
	//
	// This returns either allocated network or an error.
	Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks AllocationCallbacks) (net.IPNet, error)
}
