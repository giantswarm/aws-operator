package release

import (
	"context"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type ConditionFunc func() error

type ConditionSet interface {
	// CRDExists returns a function waiting for the CRD to appear in the
	// Kubernetes API described by the given definition.
	CRDExists(ctx context.Context, crd *apiextensionsv1beta1.CustomResourceDefinition) ConditionFunc
	// PodExists returns a function waiting for a Pod to appear in the
	// Kubernetes API described by the given label selector.
	PodExists(ctx context.Context, namespace, labelSelector string) ConditionFunc
	// SecretExists returns a function waiting for the Secret to appear in the
	// Kubernetes API described by the given name.
	SecretExists(ctx context.Context, namespace, name string) ConditionFunc
	// SecretNotExist returns a function waiting for the Secret to
	// disappear in the Kubernetes API described by the given name.
	SecretNotExist(ctx context.Context, namespace, name string) ConditionFunc
}
