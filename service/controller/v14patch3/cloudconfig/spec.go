package cloudconfig

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3/encrypter/vault"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	CloudConfigVersion = "v_3_5_1"
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

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
