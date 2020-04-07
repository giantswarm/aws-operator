package template

type ParamsMainInstance struct {
	Master ParamsMainInstanceMaster
}

type ParamsMainInstanceMaster struct {
	AZ         string
	EtcdVolume ParamsMainInstanceMasterEtcdVolume
}

type ParamsMainInstanceMasterEtcdVolume struct {
	Name string
}
