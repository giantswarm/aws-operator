package template

type ParamsMainLaunchConfiguration struct {
	BlockDeviceMapping ParamsMainLaunchConfigurationBlockDeviceMapping
	Instance           ParamsMainLaunchConfigurationInstance
	SmallCloudConfig   ParamsMainLaunchConfigurationSmallCloudConfig
}

type ParamsMainLaunchConfigurationBlockDeviceMapping struct {
	Docker  ParamsMainLaunchConfigurationBlockDeviceMappingDocker
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
	Size string
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
