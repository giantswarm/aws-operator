package template

type ParamsMainLaunchTemplate struct {
	BlockDeviceMapping    ParamsMainLaunchTemplateBlockDeviceMapping
	Instance              ParamsMainLaunchTemplateInstance
	SmallCloudConfigs     []ParamsMainLaunchTemplateSmallCloudConfig
	MasterSecurityGroupID string
	ResourceName          string
}

type ParamsMainLaunchTemplateBlockDeviceMapping struct {
	Docker  ParamsMainLaunchTemplateBlockDeviceMappingDocker
	Kubelet ParamsMainLaunchTemplateBlockDeviceMappingKubelet
	Logging ParamsMainLaunchTemplateBlockDeviceMappingLogging
}

type ParamsMainLaunchTemplateInstance struct {
	Image      string
	Monitoring bool
	Type       string
}

type ParamsMainLaunchTemplateBlockDeviceMappingDocker struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingDockerVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingDockerVolume struct {
	Size int
}

type ParamsMainLaunchTemplateBlockDeviceMappingKubelet struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingKubeletVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingKubeletVolume struct {
	Size int
}

type ParamsMainLaunchTemplateBlockDeviceMappingLogging struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingLoggingVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingLoggingVolume struct {
	Size int
}

type ParamsMainLaunchTemplateSmallCloudConfig struct {
	S3URL string
}
