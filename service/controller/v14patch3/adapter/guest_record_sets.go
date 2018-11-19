package adapter

import "github.com/giantswarm/aws-operator/service/controller/v19/key"

type GuestRecordSetsAdapter struct {
	BaseDomain                 string
	EtcdDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = baseDomain(config)
	a.EtcdDomain = key.EtcdDomain(config.CustomObject)
	a.ClusterID = clusterID(config)
	a.MasterInstanceResourceName = masterInstanceResourceName(config)
	a.Route53Enabled = route53Enabled(config)

	return nil
}
