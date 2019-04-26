package cloudconfig

import "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

// templateData is a composed data type that adds encrypter type info to
// AWSConfigSpec.
type templateData struct {
	v1alpha1.AWSConfigSpec
	EncrypterType string
	VaultAddress  string
	EncryptionKey string
}
