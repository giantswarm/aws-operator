package adapter

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

type GuestRecordSetsAdapter struct {
	BaseDomain                 string
	EtcdDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = key.ClusterBaseDomain(config.CustomObject)
	a.EtcdDomain = fmt.Sprintf("etcd.%s.%s.", key.ClusterID(config.CustomObject), key.ClusterBaseDomain(config.CustomObject))
	a.ClusterID = key.ClusterID(config.CustomObject)
	a.MasterInstanceResourceName = config.StackState.MasterInstanceResourceName
	a.Route53Enabled = config.Route53Enabled

	return nil
}
