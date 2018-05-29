package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v9patch2/key"
	"github.com/giantswarm/aws-operator/service/controller/v9patch2/templates"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v9patch2/templates/cloudformation/guest/instance.go
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
	AZ           string
	CloudConfig  string
	DockerVolume instanceAdapterMasterDockerVolume
	EtcdVolume   instanceAdapterMasterEtcdVolume
	Instance     instanceAdapterMasterInstance
}

type instanceAdapterMasterDockerVolume struct {
	Name string
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
		i.Image.ID = config.StackState.MasterImageID
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
			CloudConfigVersion: config.StackState.MasterCloudConfigVersion,
		}
		rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Master.CloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.DockerVolume.Name = key.DockerVolumeName(config.CustomObject)

		i.Master.EtcdVolume.Name = key.EtcdVolumeName(config.CustomObject)

		i.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName

		i.Master.Instance.Type = config.StackState.MasterInstanceType
	}

	return nil
}
