package tccpnencryption

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
)

const (
	name = "tccpnencryption"
)

type Config struct {
	G8sClient versioned.Interface
	Encrypter encrypter.Interface
	Logger    micrologger.Logger
}

// Resource implements the operatorkit Resource interface to fill the operator's
// controller context with an appropriate encryption key. The resource
// implementation ensures the creation of the Tenant Cluster's encryption key as
// well as its deletion. The controller context structure looks as follows.
//
//     cc.Status.TenantCluster.Encryption.Key
//
type Resource struct {
	g8sClient versioned.Interface
	encrypter encrypter.Interface
	logger    micrologger.Logger
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

	r := &Resource{
		encrypter: config.Encrypter,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}
