package statusresource

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const (
	Name = "status"
)

type ResourceConfig struct {
	BackOffFactory      func() backoff.Interface
	ClusterEndpointFunc func(v interface{}) (string, error)
	ClusterIDFunc       func(v interface{}) (string, error)
	ClusterStatusFunc   func(v interface{}) (providerv1alpha1.StatusCluster, error)
	// TODO replace this with a G8sClient to fetch the node versions from the
	// NodeConfig status once we can use the NodeConfig for general node
	// management. As of now NodeConfig CRs are still used for draining in older
	// tenant clusters.
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
	TenantCluster            tenantcluster.Interface
	VersionBundleVersionFunc func(v interface{}) (string, error)
}

type Resource struct {
	backOffFactory           func() backoff.Interface
	clusterEndpointFunc      func(v interface{}) (string, error)
	clusterIDFunc            func(v interface{}) (string, error)
	clusterStatusFunc        func(v interface{}) (providerv1alpha1.StatusCluster, error)
	k8sClient                k8sclient.Interface
	logger                   micrologger.Logger
	nodeCountFunc            func(v interface{}) (int, error)
	restClient               rest.Interface
	tenantCluster            tenantcluster.Interface
	versionBundleVersionFunc func(v interface{}) (string, error)
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.BackOffFactory == nil {
		config.BackOffFactory = func() backoff.Interface { return backoff.NewMaxRetries(3, 1*time.Second) }
	}
	if config.ClusterEndpointFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterEndpointFunc must not be empty", config)
	}
	if config.ClusterIDFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIDFunc must not be empty", config)
	}
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
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}
	if config.VersionBundleVersionFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.VersionBundleVersionFunc must not be empty", config)
	}

	r := &Resource{
		backOffFactory:           config.BackOffFactory,
		clusterEndpointFunc:      config.ClusterEndpointFunc,
		clusterIDFunc:            config.ClusterIDFunc,
		clusterStatusFunc:        config.ClusterStatusFunc,
		logger:                   config.Logger,
		nodeCountFunc:            config.NodeCountFunc,
		restClient:               config.RESTClient,
		tenantCluster:            config.TenantCluster,
		versionBundleVersionFunc: config.VersionBundleVersionFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) applyPatches(ctx context.Context, accessor metav1.Object, patches []Patch) error {
	patches = append(patches, Patch{
		Op:    "test",
		Value: accessor.GetResourceVersion(),
		Path:  "/metadata/resourceVersion",
	})

	b, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}
	p := ensureSelfLink(accessor.GetSelfLink())

	err = r.restClient.Patch(types.JSONPatchType).AbsPath(p).Body(b).Do().Error()
	if errors.IsConflict(err) {
		return microerror.Mask(err)
	} else if errors.IsResourceExpired(err) {
		return microerror.Mask(err)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func ensureDefaultPatches(clusterStatus providerv1alpha1.StatusCluster, patches []Patch) []Patch {
	conditionsEmpty := clusterStatus.Conditions == nil
	nodesEmpty := clusterStatus.Nodes == nil
	versionsEmpty := clusterStatus.Versions == nil

	if conditionsEmpty && nodesEmpty && versionsEmpty {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/status",
			Value: Status{},
		})
	}

	if conditionsEmpty {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/status/cluster/conditions",
			Value: []providerv1alpha1.StatusClusterCondition{},
		})
	}

	if nodesEmpty {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/status/cluster/nodes",
			Value: []providerv1alpha1.StatusClusterNode{},
		})
	}

	if versionsEmpty {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/status/cluster/versions",
			Value: []providerv1alpha1.StatusClusterVersion{},
		})
	}

	return patches
}

func ensureSelfLink(p string) string {
	if strings.HasSuffix(p, "/status") {
		return p
	}

	return p + "/status"
}
