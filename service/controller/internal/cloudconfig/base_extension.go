package cloudconfig

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter/vault"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type baseExtension struct {
	registryDomain string
	cluster        infrastructurev1alpha2.AWSCluster
	encrypter      encrypter.Interface
	encryptionKey  string
}

func (e *baseExtension) templateData() templateData {
	var encrypterType string
	var vaultAddress string

	v, ok := e.encrypter.(*vault.Encrypter)
	if ok {
		encrypterType = encrypter.VaultBackend
		vaultAddress = v.Address()
	} else {
		encrypterType = encrypter.KMSBackend
	}

	awsRegion := key.Region(e.cluster)

	data := templateData{
		AWSRegion:      awsRegion,
		EncryptionKey:  e.encryptionKey,
		EncrypterType:  encrypterType,
		IsChinaRegion:  key.IsChinaRegion(awsRegion),
		RegistryDomain: e.registryDomain,
		VaultAddress:   vaultAddress,
	}

	return data
}

func (e *baseExtension) encrypt(ctx context.Context, data []byte) ([]byte, error) {
	var encrypted []byte
	{
		e, err := e.encrypter.Encrypt(ctx, e.encryptionKey, string(data))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		encrypted = []byte(e)
	}

	return encrypted, nil
}
