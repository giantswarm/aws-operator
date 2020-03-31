package cloudconfig

import (
	"context"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type baseExtension struct {
	cluster        infrastructurev1alpha2.AWSCluster
	encrypter      encrypter.Interface
	encryptionKey  string
	masterSubnet   net.IPNet
	masterID       int
	registryDomain string
}

func (e *baseExtension) templateDataTCCPN() TemplateData {
	awsRegion := key.Region(e.cluster)

	data := TemplateData{
		AWSRegion:            awsRegion,
		IsChinaRegion:        key.IsChinaRegion(awsRegion),
		MasterENIAddress:     key.ControlPlaneENIIpAddress(e.masterSubnet),
		MasterENIGateway:     key.ControlPlaneENIGateway(e.masterSubnet),
		MasterENISubnetSize:  key.ControlPlaneENISubnetSize(e.masterSubnet),
		MasterENIName:        key.ControlPlaneENIName(&e.cluster, e.masterID),
		MasterEtcdVolumeName: key.ControlPlaneVolumeNameEtcd(&e.cluster, e.masterID),
		RegistryDomain:       e.registryDomain,
	}

	return data
}

func (e *baseExtension) templateDataTCCP() TemplateData {
	awsRegion := key.Region(e.cluster)

	data := TemplateData{
		AWSRegion:      awsRegion,
		IsChinaRegion:  key.IsChinaRegion(awsRegion),
		RegistryDomain: e.registryDomain,
	}

	return data
}

func (e *baseExtension) templateDataTCNP() TemplateData {
	awsRegion := key.Region(e.cluster)

	data := TemplateData{
		AWSRegion:      awsRegion,
		IsChinaRegion:  key.IsChinaRegion(awsRegion),
		RegistryDomain: e.registryDomain,
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
