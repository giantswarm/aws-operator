package update

import (
	"context"
)

type Interface interface {
	// Test executes the update test using the configured provider implementation.
	// The provider implementation has to be aware of the guest cluster it has to
	// act against. The test processes the following steps to ensure updates
	// work.
	//
	//     - Lookup the current version bundle version.
	//     - Lookup the next version bundle version.
	//     - Stop the test if there is no current version bundle version.
	//     - Stop the test if there is no next version bundle version.
	//     - Update the guest cluster to next version.
	//     - Wait for the guest cluster to be completely updated.
	//
	Test(ctx context.Context) error
}
