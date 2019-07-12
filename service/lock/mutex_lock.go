package lock

import (
	"context"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type MutexLockConfig struct {
	Logger micrologger.Logger
}

// MutexLock implements Interface using sync.Mutex. For now we use a shared
// instance of *MutexLock for all IPAM related activity of network packages in
// the legacy controllers and ipam resources in the clusterapi controllers.
type MutexLock struct {
	logger micrologger.Logger

	mutex sync.Mutex
}

func New(config MutexLockConfig) (*MutexLock, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	l := &MutexLock{
		logger: config.Logger,

		mutex: sync.Mutex{},
	}

	return l, nil
}

func (l *MutexLock) Lock(ctx context.Context) error {
	l.mutex.Lock()
	return nil
}

func (l *MutexLock) Unlock(ctx context.Context) error {
	l.mutex.Unlock()
	return nil
}
