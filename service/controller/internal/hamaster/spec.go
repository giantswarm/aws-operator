package hamaster

import "context"

// is not thread safe

// Here we try to find the subnet of the master node which is associated to a
// specific availability zone. It is not possible right now to run 3 masters
// in 1 or 2 availability zones. The system is limited to the following two
// scenarios.
//
//     * 1 master, 1 availability zone
//     * 3 master, 3 availability zone
//

type Interface interface {
	AZ() string
	ID() int
	Init(ctx context.Context, cluster string) error
	Next()
	Reconciled() bool
}
