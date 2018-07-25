package statusresource

import (
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/guestcluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/rest"
)

const (
	Name = "status"
)

type Config struct {
	ClusterEndpointFunc func(v interface{}) (string, error)
	ClusterIDFunc       func(v interface{}) (string, error)
	ClusterStatusFunc   func(v interface{}) (providerv1alpha1.StatusCluster, error)
	// TODO replace this with a G8sClient to fetch the node versions from the
	// NodeConfig status once we can use the NodeConfig for general node
	// management. As of now NodeConfig CRs are still used for draining in older
	// guest clusters.
	GuestCluster  guestcluster.Interface
	Logger        micrologger.Logger
	NodeCountFunc func(v interface{}) (int, error)
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
	VersionBundleVersionFunc func(v interface{}) (string, error)
}

type Resource struct {
	clusterEndpointFunc      func(v interface{}) (string, error)
	clusterIDFunc            func(v interface{}) (string, error)
	clusterStatusFunc        func(v interface{}) (providerv1alpha1.StatusCluster, error)
	guestCluster             guestcluster.Interface
	logger                   micrologger.Logger
	nodeCountFunc            func(v interface{}) (int, error)
	restClient               rest.Interface
	versionBundleVersionFunc func(v interface{}) (string, error)
}

func New(config Config) (*Resource, error) {
	if config.ClusterEndpointFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterEndpointFunc must not be empty", config)
	}
	if config.ClusterIDFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIDFunc must not be empty", config)
	}
	if config.ClusterStatusFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterStatusFunc must not be empty", config)
	}
	if config.GuestCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestCluster must not be empty", config)
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
		clusterEndpointFunc:      config.ClusterEndpointFunc,
		clusterIDFunc:            config.ClusterIDFunc,
		clusterStatusFunc:        config.ClusterStatusFunc,
		guestCluster:             config.GuestCluster,
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
