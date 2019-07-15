package locker

import (
	"context"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type MutexLockerConfig struct {
	Logger micrologger.Logger
}

// MutexLocker implements Interface using sync.Mutex. For now we use a shared
// instance of *MutexLocker for all IPAM related activity of network packages in
// the legacy controllers and ipam resources in the clusterapi controllers.
type MutexLocker struct {
	logger micrologger.Logger

	mutex sync.Mutex
}

func NewMutexLocker(config MutexLockerConfig) (*MutexLocker, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	l := &MutexLocker{
		logger: config.Logger,

		mutex: sync.Mutex{},
	}

	return l, nil
}

func (l *MutexLocker) Lock(ctx context.Context) error {
	l.mutex.Lock()
	return nil
}

func (l *MutexLocker) Unlock(ctx context.Context) error {
	l.mutex.Unlock()
	return nil
}
