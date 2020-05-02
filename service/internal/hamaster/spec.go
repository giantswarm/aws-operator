package hamaster

import "context"

// Interface describes implementation details of the state machine. Note that
// implementations may not be thread safe and therefore cannot be used in
// parallel by multiple goroutines.
type Interface interface {
	// AZ returns the master availability zones. Consecutive calls to AZ cycle
	// through the availability zones over and over again. Given 2 availability
	// zones A and B in a HA Masters setup of 3 masters, calling AZ until
	// Reconciled returns true results in availability zones A, B and A again
	// being returned.
	AZ() string
	// ID returns the Master ID which can either be 0, 1, 2 or 3. Master ID 0 is
	// constantly returned in a single master setup. Consecutive calls to ID in a
	// HA Masters setup return 1, 2 and then 3. Once 3 got returned the next call
	// to ID starts over at 1 again.
	ID() int
	// Init starts the initialization phase and fetches the AWSCluster and
	// AWSControlPlane CRs using the cluster ID label. This allocates the state
	// machine's information mappings. Once Init got successfully executed the
	// state machine can be used for reconciliation.
	Init(ctx context.Context, obj interface{}) error
	// Next pushes the current pointer of the state machine forward so that it can
	// cycle through the information mappings it allocated during the
	// initialization phase.
	Next()
	// Reconciled returns false upon initialization until consecutive calls to
	// Next caused the state machine to cycle through the internal information
	// mappings. Once a cycle is completed Reconciled returns true and the next
	// cycle can start over again.
	Reconciled() bool
}
