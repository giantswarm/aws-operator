package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/templates"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v10/templates/cloudformation/guest/instance.go
//

type instanceAdapter struct {
	Cluster instanceAdapterCluster
	Image   instanceAdapterImage
	Master  instanceAdapterMaster
}

type instanceAdapterCluster struct {
	ID string
}

type instanceAdapterImage struct {
	ID string
}

type instanceAdapterMaster struct {
	AZ          string
	CloudConfig string
	EtcdVolume  instanceAdapterMasterEtcdVolume
	Instance    instanceAdapterMasterInstance
}

type instanceAdapterMasterEtcdVolume struct {
	Name string
}

type instanceAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

func (i *instanceAdapter) Adapt(config Config) error {
	{
		i.Cluster.ID = key.ClusterID(config.CustomObject)
	}

	{
		imageID, err := key.ImageID(config.CustomObject)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Image.ID = imageID
	}

	{
		i.Master.AZ = key.AvailabilityZone(config.CustomObject)

		accountID, err := AccountID(config.Clients)
		if err != nil {
			return microerror.Mask(err)
		}
		c := SmallCloudconfigConfig{
			MachineType:        prefixMaster,
			Region:             key.Region(config.CustomObject),
			S3URI:              fmt.Sprintf("%s-g8s-%s", accountID, i.Cluster.ID),
			CloudConfigVersion: cloudconfig.MasterCloudConfigVersion,
		}
		rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Master.CloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.EtcdVolume.Name = key.EtcdVolumeName(config.CustomObject)

		i.Master.Instance.ResourceName = config.MasterInstanceResourceName

		i.Master.Instance.Type = key.MasterInstanceType(config.CustomObject)
	}

	return nil
}
