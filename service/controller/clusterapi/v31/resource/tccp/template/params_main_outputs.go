package template

type ParamsOutputs struct {
	Master          ParamsOutputsMaster
	OperatorVersion string
	Route53Enabled  bool
}

type ParamsOutputsMaster struct {
	ImageID      string
	Instance     ParamsOutputsMasterInstance
	DockerVolume ParamsOutputsMasterDockerVolume
}

type ParamsOutputsMasterInstance struct {
	ResourceName string
	Type         string
}

type ParamsOutputsMasterDockerVolume struct {
	ResourceName string
}
