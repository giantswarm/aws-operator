package template

type ParamsMainOutputs struct {
	DockerVolumeSizeGB string
	Instance           ParamsMainOutputsInstance
	OperatorVersion    string
	ReleaseVersion     string
}

type ParamsMainOutputsInstance struct {
	Image string
	Type  string
}
