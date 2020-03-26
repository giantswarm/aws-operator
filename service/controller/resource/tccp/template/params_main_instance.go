package template

type ParamsMainInstance struct {
	Cluster ParamsMainInstanceCluster
	Image   ParamsMainInstanceImage
	Master  ParamsMainInstanceMaster
}

type ParamsMainInstanceCluster struct {
	ID string
}

type ParamsMainInstanceImage struct {
	ID string
}

type ParamsMainInstanceMaster struct {
	AZ            string
	CloudConfig   string
	DockerVolume  ParamsMainInstanceMasterDockerVolume
	EtcdVolume    ParamsMainInstanceMasterEtcdVolume
	LogVolume     ParamsMainInstanceMasterLogVolume
	Instance      ParamsMainInstanceMasterInstance
	PrivateSubnet string
}

type ParamsMainInstanceMasterDockerVolume struct {
	Name         string
	ResourceName string
}

type ParamsMainInstanceMasterEtcdVolume struct {
	Name string
}

type ParamsMainInstanceMasterLogVolume struct {
	Name string
}

type ParamsMainInstanceMasterInstance struct {
	ResourceName string
	Type         string
	Monitoring   bool
}

// SmallCloudconfigConfig represents the data structure required for executing
// the small cloudconfig template.
type SmallCloudconfigConfig struct {
	S3URL string
}
