package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/cloudconfigv3"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/instance.yaml

type instanceAdapter struct {
	MasterAZ               string
	MasterImageID          string
	MasterInstanceType     string
	MasterSecurityGroupID  string
	MasterSmallCloudConfig string
}

func (i *instanceAdapter) getInstance(cfg Config) error {
	if len(cfg.CustomObject.Spec.AWS.Masters) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	i.MasterAZ = keyv2.AvailabilityZone(cfg.CustomObject)
	i.MasterImageID = keyv2.MasterImageID(cfg.CustomObject)
	i.MasterInstanceType = keyv2.MasterInstanceType(cfg.CustomObject)

	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := keyv2.ClusterID(cfg.CustomObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType:        prefixMaster,
		Region:             cfg.CustomObject.Spec.AWS.Region,
		S3URI:              s3URI,
		CloudConfigVersion: cloudconfigv3.MasterCloudConfigVersion,
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	i.MasterSmallCloudConfig = smallCloudConfig

	return nil
}
