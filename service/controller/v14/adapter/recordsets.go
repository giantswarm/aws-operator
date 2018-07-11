package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/v14/key"
	"github.com/giantswarm/microerror"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v14/templates/cloudformation/guest/recordsets.go
//

type recordSetsAdapter struct {
	BaseDomain                 string
	GuestHostedZoneNameServers string
	Route53Enabled             bool
}

func (r *recordSetsAdapter) getRecordSets(cfg Config) error {
	r.BaseDomain = key.BaseDomain(cfg.CustomObject)
	r.Route53Enabled = cfg.Route53Enabled

	return nil
}

func (r *recordSetsAdapter) getHostPostRecordSets(cfg Config) error {
	var err error

	err = r.getRecordSets(cfg)
	if err != nil {
		return microerror.Mask(err)
	}

	r.GuestHostedZoneNameServers = cfg.StackState.HostedZoneNameServers

	return nil
}
