package kubelock

import (
	"context"
	"time"
)

const (
	// DefaultTTL is default time to live for the lock.
	DefaultTTL = 5 * time.Minute
)

// Interface is the interface of a distributed Kubernetes lock. The default
// implementation is KubeLock.
//
// The typical usage for a namespace resource may look like:
//
//	kubeLock.Lock("my-lock-name").Namespace("my-namespace").Acquire(ctx, "my-configmap", kubelock.AcquireOptions{})
//
// The typical usage for a cluster scope resource may look like:
//
//	kubeLock.Lock("my-lock-name").Acquire(ctx, "my-namespace", kubelock.ReleaseOptions{})
//
type Interface interface {
	// Lock creates a lock with the given name. The name will be used to
	// create annotation prefixed with "kubelock.giantswarm.io/" on the
	// Kubernetes resource. Value of this annotation stores the lock data.
	//
	// NOTE: The name parameter is not validated but it must (together with
	// the annotation prefix mentioned) be a valid annotation key.
	Lock(name string) NamespaceableLock
}

type Lock interface {
	// Acquire tries to acquire the lock on a Kubernetes resource with the
	// given name.
	//
	// This method returns an error matched by IsAlreadyExists if the lock
	// already exists on the resource, it is not expired and it has the same
	// owner (set in options).
	//
	// This method returns an error matched by IsOwnerMismatch if the lock
	// already exists on the resource and it is not expired but was acquired
	// by a different owner (set in options).
	Acquire(ctx context.Context, name string, options AcquireOptions) error
	// Release tries to release the lock on a Kubernetes resource with the
	// given name.
	//
	// This method returns an error matched by IsNotFound if the lock
	// does not exist on the resource or it is expired.
	//
	// This method returns an error matched by IsOwnerMismatch if the lock
	// already exists and it is not expired but it was acquired by
	// a different owner (set in options).
	Release(ctx context.Context, name string, options ReleaseOptions) error
}

type NamespaceableLock interface {
	Lock

	// Namespace creates a lock that can be acquired on Kubernetes
	// resources in the given namespace.
	Namespace(ns string) Lock
}
