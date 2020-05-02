package encryptionensurer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/encrypter"
)

const (
	name = "encryptionensurer"
)

type Config struct {
	Encrypter encrypter.Interface
	Logger    micrologger.Logger
}

// Resource implements the operatorkit Resource interface to ensure an
// appropriate encryption key for the Tenant Cluster. The resource
// implementation ensures the creation of the Tenant Cluster's encryption key as
// well as its deletion. The encryptionensurer resource is reconciled upon the
// AWSCluster CR which defines the TCCP Cloud Formation stack. With the provider
// specific Cluster CR we ensure the encryption key. Note that the TCNP stack
// which is managed for Node Pools also needs the encryption key. It is fetched
// and put into the controller context by the encryptionsearcher resource.
type Resource struct {
	encrypter encrypter.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
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
