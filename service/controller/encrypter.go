package controller

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/kms"
)

func newEncrypterObject(getter encrypterConfigGetter) (encrypter.Interface, error) {
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

func newEncrypterRoleManager(getter encrypterConfigGetter) (encrypter.RoleManager, error) {
	return nil, nil
}
