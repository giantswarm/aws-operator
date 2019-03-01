package adapter

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

type HostPostRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	GuestHostedZoneNameServers string
	Route53Enabled             bool
}

func (a *HostPostRecordSetsAdapter) Adapt(ctx context.Context, config Config) error {
	a.BaseDomain = key.BaseDomain(config.CustomObject)
	a.ClusterID = key.ClusterID(config.CustomObject)
	a.Route53Enabled = config.Route53Enabled
	a.GuestHostedZoneNameServers = config.StackState.HostedZoneNameServers

	return nil
}
