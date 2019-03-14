package template

type ParamsMainOutputs struct {
	CloudConfig        ParamsMainOutputsCloudConfig
	DockerVolumeSizeGB string
	Instance           ParamsMainOutputsInstance
	VersionBundle      ParamsMainOutputsVersionBundle
}

type ParamsMainOutputsCloudConfig struct {
	Version string
}

type ParamsMainOutputsInstance struct {
	Image string
	Type  string
}

type ParamsMainOutputsVersionBundle struct {
	Version string
}
