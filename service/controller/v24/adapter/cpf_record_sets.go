package adapter

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

type CPFRecordSets struct {
	BaseDomain                 string
	ClusterID                  string
	GuestHostedZoneNameServers string
	Route53Enabled             bool
}

func (a *CPFRecordSets) Adapt(ctx context.Context, config Config) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	a.BaseDomain = key.BaseDomain(config.CustomObject)
	a.ClusterID = key.ClusterID(config.CustomObject)
	a.Route53Enabled = config.Route53Enabled
	a.GuestHostedZoneNameServers = cc.Status.Cluster.HostedZoneNameServers

	return nil
}
