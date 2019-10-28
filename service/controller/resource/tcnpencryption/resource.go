package tcnpencryption

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/encrypter"
)

const (
	name = "tcnpencryptionv31"
)

type Config struct {
	CMAClient clientset.Interface
	Encrypter encrypter.Interface
	Logger    micrologger.Logger
}

// Resource implements the operatorkit Resource interface to fill the operator's
// controller context with an appropriate encryption key. The controller context
// structure looks as follows.
//
//     cc.Status.TenantCluster.Encryption.Key
//
type Resource struct {
	cmaClient clientset.Interface
	encrypter encrypter.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		encrypter: config.Encrypter,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}
