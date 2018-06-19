package cloudconfig

import (
	"context"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter/vault"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	CloudConfigVersion = "v_3_3_3"
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
	ctx          context.Context
	customObject v1alpha1.AWSConfig
	encrypter    encrypter.Interface
}

func (e *baseExtension) templateData() TemplateData {
	var encrypterType string
	var vaultAddress string
	var encryptionKey string
	v, ok := e.encrypter.(*vault.Encrypter)
	if ok {
		encrypterType = encrypter.VaultBackend
		encryptionKey, _ = v.EncryptionKey(e.ctx, e.customObject)
		// Debug, fixed vault IP
		// vaultAddress = v.Address()
		vaultAddress = "https://172.19.4.88:8200"
	} else {
		encrypterType = encrypter.KMSBackend
	}
	data := TemplateData{
		AWSConfigSpec: e.customObject.Spec,
		EncrypterType: encrypterType,
		VaultAddress:  vaultAddress,
		EncryptionKey: encryptionKey,
	}

	return data
}

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
