package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/randomkeys"
)

// templateData is a composed data type that adds encrypter type info to
// AWSConfigSpec.
type templateData struct {
	v1alpha1.AWSConfigSpec
	EncrypterType string
	VaultAddress  string
	EncryptionKey string
}

type IgnitionTemplateData struct {
	CustomObject v1alpha1.AWSConfig
	ClusterCerts certs.Cluster
	ClusterKeys  randomkeys.Cluster
	Images       k8scloudconfig.Images
}
