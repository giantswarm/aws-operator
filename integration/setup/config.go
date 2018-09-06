package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
)

type Config struct {
	AWSClient *awsclient.Client
	Guest     *framework.Guest
	Host      *framework.Host

	Encrypter string
}

func (c Config) Validate() error {
	if c.AWSClient == nil {
		return microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", c)
	}
	if c.Guest == nil {
		return microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", c)
	}
	if c.Host == nil {
		return microerror.Maskf(invalidConfigError, "%T.Host must not be empty", c)
	}

	if c.Encrypter != "kms" && c.Encrypter != "vault" {
		return microerror.Maskf(invalidConfigError, "%T.Encrypter must be either `kms` or `vault`, got %#q", c, c.Encrypter)
	}

	return nil
}

type extendedConfig struct {
	Config
	VaultAddress string
	VPCPeerID    string
}
