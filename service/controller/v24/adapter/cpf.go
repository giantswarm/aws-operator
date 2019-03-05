package adapter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/routetable"
)

type CPFConfig struct {
	RouteTable *routetable.RouteTable

	AvailabilityZones          []v1alpha1.AWSConfigStatusAWSAvailabilityZone
	BaseDomain                 string
	ClusterID                  string
	EncrypterBackend           string
	GuestHostedZoneNameServers string
	NetworkCIDR                string
	Route53Enabled             bool
}

// CPF is the adapter collection for the Control Plane Finalizer management.
type CPF struct {
	RecordSets  *CPFRecordSets
	RouteTables *CPFRouteTables
}

func NewCPF(config CPFConfig) (*CPF, error) {
	var err error

	var recordSets *CPFRecordSets
	{
		recordSets = &CPFRecordSets{
			BaseDomain:                 config.BaseDomain,
			ClusterID:                  config.ClusterID,
			GuestHostedZoneNameServers: config.GuestHostedZoneNameServers,
			Route53Enabled:             config.Route53Enabled,
		}
	}

	var routeTables *CPFRouteTables
	{
		c := CPFRouteTablesConfig{
			RouteTable: config.RouteTable,

			AvailabilityZones: config.AvailabilityZones,
			EncrypterBackend:  config.EncrypterBackend,
			NetworkCIDR:       config.NetworkCIDR,
		}

		routeTables, err = NewCPFRouteTables(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	cpf := &CPF{
		RecordSets:  recordSets,
		RouteTables: routeTables,
	}

	return cpf, nil
}

func (a *CPF) Boot(ctx context.Context) error {
	err := a.RouteTables.Boot(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
