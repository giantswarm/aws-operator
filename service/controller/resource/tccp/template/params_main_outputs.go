package template

type ParamsMainOutputs struct {
	Master          ParamsMainOutputsMaster
	OperatorVersion string
	Route53Enabled  bool
}

type ParamsMainOutputsMaster struct {
	ImageID      string
	Instance     ParamsMainOutputsMasterInstance
	DockerVolume ParamsMainOutputsMasterDockerVolume
}

type ParamsMainOutputsMasterInstance struct {
	ResourceName string
	Type         string
}

type ParamsMainOutputsMasterDockerVolume struct {
	ResourceName string
}
