package provider

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

type Clients interface {
	// G8sClient returns a properly configured control plane client for the Giant
	// Swarm API Extensions Types.
	G8sClient() versioned.Interface
}

type Interface interface {
	AddWorker() error
	NumMasters() (int, error)
	NumWorkers() (int, error)
	RemoveWorker() error
	WaitForNodes(ctx context.Context, num int) error
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Waiter interface {
	// WaitForNodesReady waits for the given number of expected tenant cluster
	// nodes to be ready.
	WaitForNodesReady(ctx context.Context, expectedNodes int) error
}
