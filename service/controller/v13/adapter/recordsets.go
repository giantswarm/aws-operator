package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/v13/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
	"github.com/giantswarm/microerror"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v13/templates/cloudformation/guest/recordsets.go
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

	// When Route53 isn't enabled HostedZone isn't created and there are no
	// outputs.
	if !r.Route53Enabled {
		return nil
	}

	var cf *cloudformation.CloudFormation
	{
		c := cloudformation.Config{
			Client: cfg.Clients.CloudFormation,
		}

		cf, err = cloudformation.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// GuestHostedZoneNameServers
	{
		outputs, _, err := cf.DescribeOutputsAndStatus(key.MainGuestStackName(cfg.CustomObject))
		if err != nil {
			return microerror.Mask(err)
		}

		ns, err := cf.GetOutputValue(outputs, key.GuestHostedZoneNameServers)
		if err != nil {
			return microerror.Mask(err)
		}

		r.GuestHostedZoneNameServers = ns
	}

	return nil
}
