package adapter

import "github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"

type GuestRecordSetsAdapter struct {
	ClusterBaseDomain          string
	EtcdDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.ClusterBaseDomain = legacykey.ClusterBaseDomain(config.CustomObject)
	a.EtcdDomain = legacykey.EtcdDomain(config.CustomObject)
	a.ClusterID = legacykey.ClusterID(config.CustomObject)
	a.MasterInstanceResourceName = config.StackState.MasterInstanceResourceName
	a.Route53Enabled = config.Route53Enabled

	return nil
}
