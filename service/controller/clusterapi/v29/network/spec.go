package network

import (
	"context"
	"net"
)

type Callbacks struct {
	Collect Collector
	Persist Persister
}

// Allocator is an interface for IPAM implementation that manages network range
// and subnets for an installation.
type Allocator interface {
	// Allocate performs the subnet allocation from given fullRange for a CIDR
	// block with netSize. It requires callbacks for getting currently reserved
	// networks and persisting allocated networks. The callbacks are called within
	// a single lock. Given that callbacks work correctly, this is concurrent safe
	// within a single process e.g. when logic between multiple controllers are
	// allocating networks from the same pool.
	//
	// NOTE: This is NOT concurrent safe between distinct processes.
	//
	// This returns either allocated network or an error.
	Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks Callbacks) (net.IPNet, error)
}

// Checker implementation determines whether a subnet has to be allocated. This
// decision is being made based on the status of the Kubernetes runtime object
// defined by namespace and name.
type Checker interface {
	Check(ctx context.Context, namespace string, name string) (bool, error)
}

// Collector implementation must return all networks that are allocated on any
// given moment. Failing to do that will result in overlapping allocations.
type Collector func(context.Context) ([]net.IPNet, error)

// Persister must mutate shared persistent state so that on successful execution
// persisted networks are visible by Collector implementations.
type Persister func(context.Context, net.IPNet) error
