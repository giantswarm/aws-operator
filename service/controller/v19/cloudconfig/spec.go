package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"context"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v19/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v19/encrypter/vault"
	"github.com/giantswarm/microerror"
)

// TemplateData is a composed data type that adds encrypter type info to
// AWSConfigSpec.
type TemplateData struct {
	v1alpha1.AWSConfigSpec
	EncrypterType string
	VaultAddress  string
	EncryptionKey string
}

type baseExtension struct {
	customObject  v1alpha1.AWSConfig
	encrypter     encrypter.Interface
	encryptionKey string
}

func (e *baseExtension) templateData() TemplateData {
	var encrypterType string
	var vaultAddress string
	v, ok := e.encrypter.(*vault.Encrypter)
	if ok {
		encrypterType = encrypter.VaultBackend
		vaultAddress = v.Address()
	} else {
		encrypterType = encrypter.KMSBackend
	}
	data := TemplateData{
		AWSConfigSpec: e.customObject.Spec,
		EncrypterType: encrypterType,
		VaultAddress:  vaultAddress,
		EncryptionKey: e.encryptionKey,
	}

	return data
}

func (e *baseExtension) encryptAndGzip(ctx context.Context, data []byte) ([]byte, error) {
	var encrypted []byte
	{
		e, err := e.encrypter.Encrypt(ctx, e.encryptionKey, string(data))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		encrypted = []byte(e)
	}

	var gzipped []byte
	{
		buf := new(bytes.Buffer)
		w := gzip.NewWriter(buf)

		_, err := w.Write(encrypted)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = w.Close()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		gzipped = buf.Bytes()
	}

	return gzipped, nil
}

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
