package release

import (
	"context"
)

type ConditionFunc func() error

type ConditionSet interface {
	// PodExists returns a function waiting for a Pod to appear in the
	// Kubernetes API described by the given label selector.
	PodExists(ctx context.Context, namespace, labelSelector string) ConditionFunc
	// SecretExists returns a function waiting for the Secret to appear in the
	// Kubernetes API described by the given name.
	SecretExists(ctx context.Context, namespace, name string) ConditionFunc
}
