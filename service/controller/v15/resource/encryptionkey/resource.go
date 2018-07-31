package encryptionkey

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v15/encrypter"
)

const (
	name = "kmskeyv14"
)

type Config struct {
	Encrypter encrypter.Interface
	Logger    micrologger.Logger

	InstallationName string
}

type Resource struct {
	encrypter encrypter.Interface
	logger    micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		encrypter: config.Encrypter,
		logger:    config.Logger,

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}
