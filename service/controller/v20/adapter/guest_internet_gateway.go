package adapter

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/controller/v20/key"
)

type GuestInternetGatewayAdapter struct {
	ClusterID          string
	PrivateRouteTables []string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(cfg.CustomObject)

	// Since CloudFormation cannot recognize resource renaming, use non-indexed
	// resource name for first AZ to not recreate the internet gateway for legacy
	// clusters.
	a.PrivateRouteTables = []string{"PrivateRouteTable"}

	for i := 1; i < key.SpecAvailabilityZones(cfg.CustomObject); i++ {
		a.PrivateRouteTables = append(a.PrivateRouteTables, fmt.Sprintf("PrivateRouteTable%02d", i))
	}

	return nil
}
