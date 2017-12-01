package adapter

import (
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/instance.yaml

type instanceAdapter struct {
	AZ                     string
	IAMInstanceProfileName string
	ImageID                string
	InstanceType           string
	SecurityGroupID        string
	SmallCloudConfig       string
	SubnetID               string
	Tags                   map[string]string
}

func (i *instanceAdapter) getInstance(customObject awstpr.CustomObject, clients Clients) error {
	if len(customObject.Spec.AWS.Masters) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	i.AZ = key.AvailabilityZone(customObject)
	i.ImageID = key.MasterImageID(customObject)
	i.InstanceType = key.MasterInstanceType(customObject)
	i.IAMInstanceProfileName = key.InstanceProfileName(customObject, prefixMaster)

	return nil
}
