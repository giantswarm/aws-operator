package adapter

import (
	"encoding/base64"
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/templates"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type GuestInstanceAdapter struct {
	Cluster GuestInstanceAdapterCluster
	Image   GuestInstanceAdapterImage
	Master  GuestInstanceAdapterMaster
}

type GuestInstanceAdapterCluster struct {
	ID string
}

type GuestInstanceAdapterImage struct {
	ID string
}

type GuestInstanceAdapterMaster struct {
	AZ               string
	Ignition         string
	EncrypterBackend string
	DockerVolume     GuestInstanceAdapterMasterDockerVolume
	EtcdVolume       GuestInstanceAdapterMasterEtcdVolume
	LogVolume        GuestInstanceAdapterMasterLogVolume
	Instance         GuestInstanceAdapterMasterInstance
	PrivateSubnet    string
}

type GuestInstanceAdapterMasterDockerVolume struct {
	Name         string
	ResourceName string
}

type GuestInstanceAdapterMasterEtcdVolume struct {
	Name string
}

type GuestInstanceAdapterMasterLogVolume struct {
	Name string
}

type GuestInstanceAdapterMasterInstance struct {
	ResourceName string
	Type         string
	Monitoring   bool
}

func (i *GuestInstanceAdapter) Adapt(config Config) error {
	{
		i.Cluster.ID = key.ClusterID(config.CustomObject)
	}

	{
		i.Image.ID = config.StackState.MasterImageID
	}

	{
		zones := key.StatusAvailabilityZones(config.CustomObject)
		sort.Slice(zones, func(i, j int) bool {
			return zones[i].Name < zones[j].Name
		})

		if len(zones) < 1 {
			return microerror.Maskf(notFoundError, "CustomObject has no availability zones")
		}

		i.Master.AZ = zones[0].Name
		i.Master.PrivateSubnet = key.PrivateSubnetName(0)

		c := SmallCloudconfigConfig{
			InstanceRole: key.KindMaster,
			S3URL:        key.SmallCloudConfigS3URL(config.CustomObject, config.TenantClusterAccountID, key.KindMaster),
		}
		rendered, err := templates.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}

		i.Master.EncrypterBackend = config.EncrypterBackend

		i.Master.DockerVolume.Name = key.DockerVolumeName(config.CustomObject)

		i.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName

		i.Master.EtcdVolume.Name = key.EtcdVolumeName(config.CustomObject)

		i.Master.Ignition = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.LogVolume.Name = key.LogVolumeName(config.CustomObject)

		i.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName

		i.Master.Instance.Type = config.StackState.MasterInstanceType

		i.Master.Instance.Monitoring = config.StackState.MasterInstanceMonitoring
	}

	return nil
}
