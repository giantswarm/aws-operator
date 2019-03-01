package adapter

import (
	"context"

	"github.com/giantswarm/microerror"
)

// CPF is the adapter collection for the Control Plane Finalizer management.
type CPF struct {
	RecordSets  HostPostRecordSetsAdapter
	RouteTables HostPostRouteTablesAdapter
}

func NewCPF(ctx context.Context, config Config) (*CPF, error) {
	adapter := &CPF{}

	hydraters := []ContextHydrater{
		adapter.RecordSets.Adapt,
		adapter.RouteTables.Adapt,
	}

	for _, h := range hydraters {
		err := h(ctx, config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return adapter, nil
}
