package clusterstate

import (
	"context"
)

type Interface interface {
	// Test executes the master node test using the configured provider
	// implementation. The provider implementation has to be aware of the guest
	// cluster it has to act against. The test processes the following steps to
	// ensure scaling works.
	//
	//  - Install test app.
	//  - Wait for cluster to be ready.
	//  - Reboot master node.
	//  - Wait for cluster to be ready.
	//  - Replace master node.
	//  - Wait for cluster to be ready.
	//
	Test(ctx context.Context) error
}
