package template

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type ParamsRecordSets struct {
	BaseDomain                 string
	EtcdDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
	VPCRegion                  string
}
