package statusresource

import (
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/rest"
)

const (
	Name = "status"
)

type Config struct {
	ClusterStatusFunc func(obj interface{}) (providerv1alpha1.StatusCluster, error)
	Logger            micrologger.Logger
	NodeCountFunc     func(obj interface{}) (int, error)
	// RESTClient needs to be configured with a serializer capable of serializing
	// and deserializing the object which is watched by the informer. Otherwise
	// deserialization will fail when trying to manage the cluster status.
	//
	// For standard k8s object this is going to be e.g.
	//
	//     k8sClient.CoreV1().RESTClient()
	//
	// For CRs of giantswarm this is going to be e.g.
	//
	//     g8sClient.CoreV1alpha1().RESTClient()
	//
	RESTClient               rest.Interface
	VersionBundleVersionFunc func(obj interface{}) (string, error)
}

type Resource struct {
	clusterStatusFunc        func(obj interface{}) (providerv1alpha1.StatusCluster, error)
	logger                   micrologger.Logger
	nodeCountFunc            func(obj interface{}) (int, error)
	restClient               rest.Interface
	versionBundleVersionFunc func(obj interface{}) (string, error)
}

func New(config Config) (*Resource, error) {
	if config.ClusterStatusFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterStatusFunc must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NodeCountFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NodeCountFunc must not be empty", config)
	}
	if config.RESTClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RESTClient must not be empty", config)
	}
	if config.VersionBundleVersionFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VersionBundleVersionFunc must not be empty", config)
	}

	r := &Resource{
		clusterStatusFunc:        config.ClusterStatusFunc,
		logger:                   config.Logger,
		nodeCountFunc:            config.NodeCountFunc,
		restClient:               config.RESTClient,
		versionBundleVersionFunc: config.VersionBundleVersionFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
