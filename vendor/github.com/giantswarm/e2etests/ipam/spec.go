package ipam

import "context"

type Interface interface {
	// Test executes the cluster IPAM test using the configured provider
	// implementation. The test processes the following steps to
	// ensure the provider specific operator implements guest cluster IPAM
	// correctly.
	//
	//  - Create guest clusters #1, #2, #3.
	//  - Wait for guest clusters to be ready.
	//  - Verify that clusters have distinct subnets.
	//  - Terminate guest cluster #2 and immediately create guest cluster #4.
	//  - Wait for guest clusters to be deleted and created.  - Verify that
	//    clusters have distinct subnets and created cluster #4 did not receive
	//    same subnet that deleted cluster #2 had.
	//  - Delete guest clusters.
	//
	Test(ctx context.Context) error
}
