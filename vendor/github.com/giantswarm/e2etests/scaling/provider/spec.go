package provider

type Interface interface {
	AddWorker() error
	NumMasters() (int, error)
	NumWorkers() (int, error)
	RemoveWorker() error
	WaitForNodes(num int) error
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
