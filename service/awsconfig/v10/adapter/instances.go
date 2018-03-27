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
//     github.com/giantswarm/aws-operator/service/awsconfig/v10/templates/cloudformation/guest/instances.go
//

type instancesAdapter struct {
	Cluster instancesAdapterCluster
	Image   instancesAdapterImage
	Master  instancesAdapterMaster
}

type instancesAdapterCluster struct {
	ID string
}

type instancesAdapterImage struct {
	ID string
}

type instancesAdapterMaster struct {
	AZ          string
	CloudConfig string
	Instance    instancesAdapterMasterInstance
}

type instancesAdapterMasterInstance struct {
	ID   string
	Type string
}

func (i *instancesAdapter) Adapt(config Config) error {
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

		i.Master.Instance.ID = config.MasterInstanceID

		i.Master.Instance.Type = key.MasterInstanceType(config.CustomObject)
	}

	return nil
}
