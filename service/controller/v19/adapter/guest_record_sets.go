package adapter

import "github.com/giantswarm/aws-operator/service/controller/v19/key"

type GuestRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = key.BaseDomain(config.CustomObject)
	a.ClusterID = key.ClusterID(config.CustomObject)
	a.MasterInstanceResourceName = config.StackState.MasterInstanceResourceName
	a.Route53Enabled = config.Route53Enabled

	return nil
}
