package scaling

import (
	"context"
)

type Interface interface {
	// Test executes the scaling test using the configured provider
	// implementation. The provider implementation has to be aware of the guest
	// cluster it has to act against. The test processes the following steps to
	// ensure scaling works.
	//
	//     - Scale guest cluster up by one worker.
	//     - Wait for new guest cluster worker to be up.
	//     - Scale guest cluster down by one worker.
	//     - Wait for old guest cluster worker to be down.
	//
	Test(ctx context.Context) error
}
