package encryption

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/encrypter"
)

const (
	name = "encryptionv31"
)

type Config struct {
	Encrypter     encrypter.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

// Resource implements the operatorkit Resource interface to fill the operator's
// controller context with an appropriate encryption key. The controller context
// structure looks as follows.
//
//     cc.Status.TenantCluster.Encryption.Key
//
// The resource may be used by different controllers reconcoling different
// runtime objects. Therefore toClusterFunc must be configured accordingly. This
// may be a simple key function or a more complex lookup implementation.
type Resource struct {
	encrypter     encrypter.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		encrypter:     config.Encrypter,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}
