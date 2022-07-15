package template

type ParamsMainLaunchTemplate struct {
	BlockDeviceMapping ParamsMainLaunchTemplateBlockDeviceMapping
	Instance           ParamsMainLaunchTemplateInstance
	Metadata           ParamsMainLaunchTemplateMetadata
	Name               string
	ReleaseVersion     string
	SmallCloudConfig   ParamsMainLaunchTemplateSmallCloudConfig
}

type ParamsMainLaunchTemplateBlockDeviceMapping struct {
	Containerd ParamsMainLaunchTemplateBlockDeviceMappingContainerd
	Docker     ParamsMainLaunchTemplateBlockDeviceMappingDocker
	Kubelet    ParamsMainLaunchTemplateBlockDeviceMappingKubelet
	Logging    ParamsMainLaunchTemplateBlockDeviceMappingLogging
}

type ParamsMainLaunchTemplateInstance struct {
	Image      string
	Monitoring bool
	Type       string
}

type ParamsMainLaunchTemplateMetadata struct {
	HttpTokens string
}

type ParamsMainLaunchTemplateBlockDeviceMappingContainerd struct {
	Volume ParamsMainLaunchTemplateBlockDeviceMappingContainerdVolume
}

type ParamsMainLaunchTemplateBlockDeviceMappingContainerdVolume struct {
	Size string
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
	S3URL string
}
