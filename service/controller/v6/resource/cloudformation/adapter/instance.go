package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v6/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v6/key"
	"github.com/giantswarm/aws-operator/service/awsconfig/v6/templates"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v6/templates/cloudformation/guest/instance.go
//

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

	c := SmallCloudconfigConfig{
		MachineType:        prefixMaster,
		Region:             cfg.CustomObject.Spec.AWS.Region,
		S3URI:              s3URI,
		CloudConfigVersion: cloudconfig.MasterCloudConfigVersion,
	}
	rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
	if err != nil {
		return microerror.Mask(err)
	}
	i.MasterSmallCloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

	return nil
}
