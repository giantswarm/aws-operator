package encryptionsearcher

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/encrypter"
)

const (
	name = "encryptionsearcher"
)

type Config struct {
	G8sClient     versioned.Interface
	Encrypter     encrypter.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

// Resource implements the operatorkit Resource interface to fill the operator's
// controller context with an appropriate encryption key, assuming it got
// already created. The encryption key is created by the encryptionensurer
// resource within the cluster controller. See the controller context structure
// below.
//
//     cc.Status.TenantCluster.Encryption.Key
//
type Resource struct {
	g8sClient     versioned.Interface
	encrypter     encrypter.Interface
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
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
		g8sClient:     config.G8sClient,
		encrypter:     config.Encrypter,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}
