package kubelock

import (
	"time"
)

type AcquireOptions struct {
	// Owner is an arbitrary string representing owner of the lock.
	Owner string
	// TTL is time to live for the lock.
	TTL time.Duration
}

type ReleaseOptions struct {
	// Owner is an arbitrary string representing owner of the lock.
	Owner string
}

type lockData struct {
	Owner     string        `json:"owner"`
	CreatedAt time.Time     `json:"createdAt"`
	TTL       time.Duration `json:"ttl"`
}
