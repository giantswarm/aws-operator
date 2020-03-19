package cloudconfig

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type baseExtension struct {
	awsConfigSpec  g8sv1alpha1.AWSConfigSpec
	registryDomain string
	cluster        infrastructurev1alpha2.AWSCluster
	encrypter      encrypter.Interface
	encryptionKey  string
}

func (e *baseExtension) templateData() TemplateData {
	awsRegion := key.Region(e.cluster)
	data := TemplateData{
		AWSRegion:           awsRegion,
		AWSConfigSpec:       e.awsConfigSpec,
		IsChinaRegion:       key.IsChinaRegion(awsRegion),
		MasterENIAddresses:  []string{"TODO"},
		MasterENIGateways:   []string{"TODO"},
		MasterENISubnetSize: "TODO",
		MasterID:            0,
		RegistryDomain:      e.registryDomain,
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
