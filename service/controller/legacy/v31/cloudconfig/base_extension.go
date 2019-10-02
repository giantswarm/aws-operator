package cloudconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/encrypter/vault"
)

type baseExtension struct {
	customObject  v1alpha1.AWSConfig
	encrypter     encrypter.Interface
	encryptionKey string
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
	data := templateData{
		AWSConfigSpec: e.customObject.Spec,
		EncrypterType: encrypterType,
		VaultAddress:  vaultAddress,
		EncryptionKey: e.encryptionKey,
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
