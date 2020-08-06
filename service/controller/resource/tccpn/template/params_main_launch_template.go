package template

type ParamsMainLaunchTemplate struct {
	List []ParamsMainLaunchTemplateItem
}

type ParamsMainLaunchTemplateItem struct {
	BlockDeviceMapping    ParamsMainLaunchTemplateItemBlockDeviceMapping
	Instance              ParamsMainLaunchTemplateItemInstance
	SmallCloudConfig      ParamsMainLaunchTemplateItemSmallCloudConfig
	MasterSecurityGroupID string
	Name                  string
	Resource              string
}

type ParamsMainLaunchTemplateItemBlockDeviceMapping struct {
	Docker  ParamsMainLaunchTemplateItemBlockDeviceMappingDocker
	Kubelet ParamsMainLaunchTemplateItemBlockDeviceMappingKubelet
	Logging ParamsMainLaunchTemplateItemBlockDeviceMappingLogging
}

type ParamsMainLaunchTemplateItemInstance struct {
	Image      string
	Monitoring bool
	Type       string
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingDocker struct {
	Volume ParamsMainLaunchTemplateItemBlockDeviceMappingDockerVolume
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingDockerVolume struct {
	Size int
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingKubelet struct {
	Volume ParamsMainLaunchTemplateItemBlockDeviceMappingKubeletVolume
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingKubeletVolume struct {
	Size int
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingLogging struct {
	Volume ParamsMainLaunchTemplateItemBlockDeviceMappingLoggingVolume
}

type ParamsMainLaunchTemplateItemBlockDeviceMappingLoggingVolume struct {
	Size int
}

type ParamsMainLaunchTemplateItemSmallCloudConfig struct {
	Hash  string
	S3URL string
}
