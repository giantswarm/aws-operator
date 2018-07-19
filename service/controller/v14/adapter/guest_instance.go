package adapter

import (
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14/key"
	"github.com/giantswarm/aws-operator/service/controller/v14/templates"
)

type guestInstanceAdapter struct {
	Cluster guestInstanceAdapterCluster
	Image   guestInstanceAdapterImage
	Master  guestInstanceAdapterMaster
}

type guestInstanceAdapterCluster struct {
	ID string
}

type guestInstanceAdapterImage struct {
	ID string
}

type guestInstanceAdapterMaster struct {
	AZ               string
	CloudConfig      string
	EncrypterBackend string
	DockerVolume     guestInstanceAdapterMasterDockerVolume
	EtcdVolume       guestInstanceAdapterMasterEtcdVolume
	Instance         guestInstanceAdapterMasterInstance
}

type guestInstanceAdapterMasterDockerVolume struct {
	Name         string
	ResourceName string
}

type guestInstanceAdapterMasterEtcdVolume struct {
	Name string
}

type guestInstanceAdapterMasterInstance struct {
	ResourceName string
	Type         string
	Monitoring   bool
}

func (i *guestInstanceAdapter) Adapt(config Config) error {
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
			MachineType:             prefixMaster,
			Region:                  key.Region(config.CustomObject),
			S3Domain:                key.S3ServiceDomain(config.CustomObject),
			S3URI:                   fmt.Sprintf("%s-g8s-%s", accountID, i.Cluster.ID),
			CloudConfigVersion:      config.StackState.MasterCloudConfigVersion,
			AWSCliContainerRegistry: key.AWSCliContainerRegistry(config.CustomObject),
		}
		rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Master.CloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.EncrypterBackend = config.EncrypterBackend

		i.Master.DockerVolume.Name = key.DockerVolumeName(config.CustomObject)

		i.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName

		i.Master.EtcdVolume.Name = key.EtcdVolumeName(config.CustomObject)

		i.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName

		i.Master.Instance.Type = config.StackState.MasterInstanceType

		i.Master.Instance.Monitoring = config.StackState.MasterInstanceMonitoring
	}

	return nil
}
