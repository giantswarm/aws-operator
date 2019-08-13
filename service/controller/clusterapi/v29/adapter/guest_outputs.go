package adapter

type GuestOutputsAdapter struct {
	Master         GuestOutputsAdapterMaster
	Route53Enabled bool
	VersionBundle  GuestOutputsAdapterVersionBundle
}

func (a *GuestOutputsAdapter) Adapt(config Config) error {
	a.Route53Enabled = config.Route53Enabled
	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType
	a.Master.CloudConfig.Version = config.StackState.MasterCloudConfigVersion

	a.VersionBundle.Version = config.StackState.VersionBundleVersion

	return nil
}

type GuestOutputsAdapterMaster struct {
	ImageID      string
	Instance     GuestOutputsAdapterMasterInstance
	CloudConfig  GuestOutputsAdapterMasterCloudConfig
	DockerVolume GuestOutputsAdapterMasterDockerVolume
}

type GuestOutputsAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

type GuestOutputsAdapterMasterCloudConfig struct {
	Version string
}

type GuestOutputsAdapterMasterDockerVolume struct {
	ResourceName string
}

type GuestOutputsAdapterVersionBundle struct {
	Version string
}
