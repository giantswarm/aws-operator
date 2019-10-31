package controller

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/vault"
)

func newEncrypterObject(getter encrypterConfigGetter) (encrypter.Interface, error) {
	if getter.GetEncrypterBackend() == encrypter.VaultBackend {
		c := &vault.EncrypterConfig{
			Logger: getter.GetLogger(),

			Address: getter.GetVaultAddress(),
		}

		encrypterObject, err := vault.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return encrypterObject, nil
	}

	if getter.GetEncrypterBackend() == encrypter.KMSBackend {
		c := &kms.EncrypterConfig{
			Logger: getter.GetLogger(),

			InstallationName: getter.GetInstallationName(),
		}

		encrypterObject, err := kms.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return encrypterObject, nil
	}

	return nil, microerror.Maskf(invalidConfigError, "unknown encrypter backend %q", getter.GetEncrypterBackend())
}

func newEncrypterRoleManager(getter encrypterConfigGetter) (encrypter.RoleManager, error) {
	if getter.GetEncrypterBackend() == encrypter.VaultBackend {
		c := &vault.EncrypterConfig{
			Logger: getter.GetLogger(),

			Address: getter.GetVaultAddress(),
		}

		encrypterObject, err := vault.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return encrypterObject, nil
	}

	if getter.GetEncrypterBackend() == encrypter.KMSBackend {
		return nil, nil
	}

	return nil, microerror.Maskf(invalidConfigError, "unknown encrypter backend %q", getter.GetEncrypterBackend())
}
