package template

type ParamsMainLaunchTemplate struct {
	List []ParamsMainLaunchTemplateItem
}

type ParamsMainLaunchTemplateItem struct {
	BlockDeviceMapping    ParamsMainLaunchTemplateItemBlockDeviceMapping
	Instance              ParamsMainLaunchTemplateItemInstance
	Metadata              ParamsMainLaunchTemplateMetadata
	SmallCloudConfig      ParamsMainLaunchTemplateItemSmallCloudConfig
	MasterSecurityGroupID string
	Name                  string
	Resource              string
	ReleaseVersion        string
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

type ParamsMainLaunchTemplateMetadata struct {
	HttpTokens string
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
	S3URL string
}
