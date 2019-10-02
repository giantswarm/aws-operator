package adapter

import "github.com/giantswarm/aws-operator/service/controller/legacy/v31/key"

type GuestRecordSetsAdapter struct {
	BaseDomain                 string
	EtcdDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
	VPCRegion                  string
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = key.BaseDomain(config.CustomObject)
	a.EtcdDomain = key.EtcdDomain(config.CustomObject)
	a.ClusterID = key.ClusterID(config.CustomObject)
	a.MasterInstanceResourceName = config.StackState.MasterInstanceResourceName
	a.Route53Enabled = config.Route53Enabled
	a.VPCRegion = key.Region(config.CustomObject)

	return nil
}
