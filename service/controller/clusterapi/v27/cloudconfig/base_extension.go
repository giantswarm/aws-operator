package cloudconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter/vault"
)

type baseExtension struct {
	cluster       v1alpha1.Cluster
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
		AWSConfigSpec: cmaClusterToG8sConfig(e.cluster),
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
