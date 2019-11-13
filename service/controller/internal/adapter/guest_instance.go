package adapter

import (
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/template"
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
	CloudConfig      string
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

// SmallCloudconfigConfig represents the data structure required for executing
// the small cloudconfig template.
type SmallCloudconfigConfig struct {
	IgnitionSHA512 string
	S3URL          string
}

func (i *GuestInstanceAdapter) Adapt(config Config) error {
	{
		i.Cluster.ID = key.ClusterID(&config.CustomObject)
	}

	{
		i.Image.ID = config.StackState.MasterImageID
	}

	{
		zones := config.TenantClusterAvailabilityZones

		sort.Slice(zones, func(i, j int) bool {
			return zones[i].Name < zones[j].Name
		})

		if len(zones) < 1 {
			return microerror.Maskf(notFoundError, "CustomObject has no availability zones")
		}

		i.Master.AZ = key.MasterAvailabilityZone(config.CustomObject)
		i.Master.PrivateSubnet = key.SanitizeCFResourceName(key.PrivateSubnetName(i.Master.AZ))

		c := SmallCloudconfigConfig{
			S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&config.CustomObject, config.TenantClusterAccountID), key.S3ObjectPathTCCP(&config.CustomObject)),
		}
		rendered, err := template.Render(key.CloudConfigSmallTemplates(), c)
		if err != nil {
			return microerror.Mask(err)
		}
		i.Master.CloudConfig = base64.StdEncoding.EncodeToString([]byte(rendered))

		i.Master.EncrypterBackend = config.EncrypterBackend
		i.Master.DockerVolume.Name = key.VolumeNameDocker(config.CustomObject)
		i.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
		i.Master.EtcdVolume.Name = key.VolumeNameEtcd(config.CustomObject)
		i.Master.LogVolume.Name = key.VolumeNameLog(config.CustomObject)
		i.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
		i.Master.Instance.Type = config.StackState.MasterInstanceType
		i.Master.Instance.Monitoring = config.StackState.MasterInstanceMonitoring
	}

	return nil
}
