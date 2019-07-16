package locker

import "context"

// Interface is some form of lock implementation like achieved for in process
// locking using sync.Mutex.
type Interface interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
