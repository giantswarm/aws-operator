package template

type ParamsMainLaunchConfiguration struct {
	BlockDeviceMapping    ParamsMainLaunchConfigurationBlockDeviceMapping
	Instance              ParamsMainLaunchConfigurationInstance
	SmallCloudConfig      ParamsMainLaunchConfigurationSmallCloudConfig
	MasterSecurityGroupID string
}

type ParamsMainLaunchConfigurationBlockDeviceMapping struct {
	Docker  ParamsMainLaunchConfigurationBlockDeviceMappingDocker
	Kubelet ParamsMainLaunchConfigurationBlockDeviceMappingKubelet
	Logging ParamsMainLaunchConfigurationBlockDeviceMappingLogging
}

type ParamsMainLaunchConfigurationInstance struct {
	Image      string
	Monitoring bool
	Type       string
}

type ParamsMainLaunchConfigurationBlockDeviceMappingDocker struct {
	Volume ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume
}

type ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume struct {
	Size int
}

type ParamsMainLaunchConfigurationBlockDeviceMappingKubelet struct {
	Volume ParamsMainLaunchConfigurationBlockDeviceMappingKubeletVolume
}

type ParamsMainLaunchConfigurationBlockDeviceMappingKubeletVolume struct {
	Size int
}

type ParamsMainLaunchConfigurationBlockDeviceMappingLogging struct {
	Volume ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume
}

type ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume struct {
	Size int
}

type ParamsMainLaunchConfigurationSmallCloudConfig struct {
	S3URL string
}
