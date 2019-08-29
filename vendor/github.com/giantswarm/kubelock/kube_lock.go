package kubelock

import (
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	DynClient dynamic.Interface

	// GVR defines a resource that the lock will be
	// created on with use of the dynamic Kubernetes client. This object
	// will be passed directly in the dynamic client API calls.
	//
	// E.g. for Namespace resource it can be instantiated like:
	//
	//	schema.GroupVersionResource{
	//		Group:    "",
	//		Version:  "v1",
	//		Resource: "namespaces",
	//	}
	//
	// E.g. for CR defined by a CRD defined in
	// https://github.com/giantswarm/apiextensions/ repository it can be
	// instantiated like:
	//
	//	schema.GroupVersionResource{
	//		Group:    v1alpha1.NewAWSConfigCRD().Spec.Group,
	//		Version:  v1alpha1.NewAWSConfigCRD().Spec.Version,
	//		Resource: v1alpha1.NewAWSConfigCRD().Spec.Names.Plural,
	//	}
	GVR schema.GroupVersionResource
}

type KubeLock struct {
	dynClient dynamic.Interface

	gvr schema.GroupVersionResource
}

func New(config Config) (*KubeLock, error) {
	if config.DynClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.DynClient must not be empty", config)
	}

	if config.GVR.Version == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GVR.Version must not be empty", config)
	}
	if config.GVR.Resource == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GVR.Resource must not be empty", config)
	}

	k := &KubeLock{
		dynClient: config.DynClient,

		gvr: config.GVR,
	}

	return k, nil
}

func (k *KubeLock) Lock(name string) NamespaceableLock {
	return &namespaceableLock{
		resource: k.dynClient.Resource(k.gvr),

		lockName: name,
	}
}
