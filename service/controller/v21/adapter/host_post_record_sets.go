package adapter

import "github.com/giantswarm/aws-operator/service/controller/v21/key"

type HostPostRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	GuestHostedZoneNameServers string
	Route53Enabled             bool
}

func (r *HostPostRecordSetsAdapter) Adapt(config Config) error {
	r.BaseDomain = key.BaseDomain(config.CustomObject)
	r.ClusterID = key.ClusterID(config.CustomObject)
	r.Route53Enabled = config.Route53Enabled
	r.GuestHostedZoneNameServers = config.StackState.HostedZoneNameServers

	return nil
}
