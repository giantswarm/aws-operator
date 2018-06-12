package provider

type Interface interface {
	AddWorker() error
	NumMasters() (int, error)
	NumWorkers() (int, error)
	RemoveWorker() error
	WaitForNodesUp(num int) error
}
