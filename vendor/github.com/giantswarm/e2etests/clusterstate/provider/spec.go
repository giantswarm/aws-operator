package provider

type Interface interface {
	RebootMaster() error
	ReplaceMaster() error
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
