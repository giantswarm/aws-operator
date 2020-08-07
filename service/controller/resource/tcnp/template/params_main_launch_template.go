package template

type ParamsMainLaunchTemplate struct {
	BlockDeviceMapping ParamsMainLaunchTemplateBlockDeviceMapping
	Instance           ParamsMainLaunchTemplateInstance
	Name               string
	SmallCloudConfig   ParamsMainLaunchTemplateSmallCloudConfig
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
	Size string
}

type ParamsMainLaunchTemplateBlockDeviceMappingKubelet struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingKubeletVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingKubeletVolume struct {
	Size string
}

type ParamsMainLaunchTemplateBlockDeviceMappingLogging struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingLoggingVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingLoggingVolume struct {
	Size int
}

type ParamsMainLaunchTemplateSmallCloudConfig struct {
	Hash  string
	S3URL string
}
