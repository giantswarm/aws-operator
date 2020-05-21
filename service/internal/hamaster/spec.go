package hamaster

import "context"

type Mapping struct {
	// AZ is the master availability zones. Given 2 availability zones A and B in
	// a HA Masters setup of 3 masters, AZ will be A, B and A again in the list of
	// mappings computed by implementations of Interface.
	AZ string
	// ID can either be 0, 1, 2 or 3. Master ID 0 is omnipresent in a single
	// master setup. In a HA Masters setup ID will be 1, 2 and then 3 in the list
	// of mappings computed by implementations of Interface.
	ID int
}

type Interface interface {
	// Enabled returns true in case HA Masters is enabled. Right now this means to
	// have 3 replicas configured for the master setup.
	Enabled(ctx context.Context, obj interface{}) (bool, error)
	// Mapping fetches the AWSCluster and AWSControlPlane CRs using the cluster ID
	// label obj must provide as meta object. See the godoc of Mapping for more
	// information on the returned list of mapped information.
	Mapping(ctx context.Context, obj interface{}) ([]Mapping, error)
	// Replicas fetches the G8sControlPlane CR and returns the number of replicas
	// configured for the master setup.
	Replicas(ctx context.Context, obj interface{}) (int, error)
}
