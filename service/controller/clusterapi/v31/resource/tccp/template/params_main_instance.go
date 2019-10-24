package template

type ParamsInstance struct {
	Cluster ParamsInstanceCluster
	Image   ParamsInstanceImage
	Master  ParamsInstanceMaster
}

type ParamsInstanceCluster struct {
	ID string
}

type ParamsInstanceImage struct {
	ID string
}

type ParamsInstanceMaster struct {
	AZ               string
	CloudConfig      string
	EncrypterBackend string
	DockerVolume     ParamsInstanceMasterDockerVolume
	EtcdVolume       ParamsInstanceMasterEtcdVolume
	LogVolume        ParamsInstanceMasterLogVolume
	Instance         ParamsInstanceMasterInstance
	PrivateSubnet    string
}

type ParamsInstanceMasterDockerVolume struct {
	Name         string
	ResourceName string
}

type ParamsInstanceMasterEtcdVolume struct {
	Name string
}

type ParamsInstanceMasterLogVolume struct {
	Name string
}

type GuestInstanceMasterInstance struct {
	ResourceName string
	Type         string
	Monitoring   bool
}

// SmallCloudconfigConfig represents the data structure required for executing
// the small cloudconfig template.
type SmallCloudconfigConfig struct {
	S3URL string
}
