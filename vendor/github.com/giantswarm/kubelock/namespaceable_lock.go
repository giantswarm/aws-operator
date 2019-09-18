package kubelock

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/client-go/dynamic"
)

type namespaceableLock struct {
	resource dynamic.NamespaceableResourceInterface

	lockName string
}

func (l *namespaceableLock) Acquire(ctx context.Context, name string, options AcquireOptions) error {
	underlying := &lock{
		resource: l.resource,

		lockName: l.lockName,
	}

	err := underlying.Acquire(ctx, name, options)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (l *namespaceableLock) Namespace(ns string) Lock {
	return &lock{
		resource: l.resource.Namespace(ns),

		lockName: l.lockName,
	}
}

func (l *namespaceableLock) Release(ctx context.Context, name string, options ReleaseOptions) error {
	underlying := &lock{
		resource: l.resource,

		lockName: l.lockName,
	}

	err := underlying.Release(ctx, name, options)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
