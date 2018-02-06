package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
	// NOTE(PK): This import is disturbing. I'm not bothering. It's first candidate to go away.
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/cloudconfig"
)

// template related to this adapter: service/templates/cloudformation/guest/instance.yaml

type instanceAdapter struct {
	MasterAZ               string
	MasterInstanceType     string
	MasterSecurityGroupID  string
	MasterSmallCloudConfig string
}

func (i *instanceAdapter) getInstance(cfg Config) error {
	if len(cfg.CustomObject.Spec.AWS.Masters) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	i.MasterAZ = key.AvailabilityZone(cfg.CustomObject)
	i.MasterInstanceType = key.MasterInstanceType(cfg.CustomObject)

	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(cfg.CustomObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType:        prefixMaster,
		Region:             cfg.CustomObject.Spec.AWS.Region,
		S3URI:              s3URI,
		CloudConfigVersion: cloudconfig.MasterCloudConfigVersion,
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	i.MasterSmallCloudConfig = smallCloudConfig

	return nil
}
