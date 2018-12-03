package encryptionkey

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v17/encrypter"
)

const (
	name = "encryptionkeyv17"
)

type Config struct {
	Encrypter encrypter.Resource
	Logger    micrologger.Logger
}

type Resource struct {
	encrypter encrypter.Resource
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
