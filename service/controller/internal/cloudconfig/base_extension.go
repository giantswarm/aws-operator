package cloudconfig

import (
	"context"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type baseExtension struct {
	awsConfigSpec  g8sv1alpha1.AWSConfigSpec
	cluster        infrastructurev1alpha2.AWSCluster
	encrypter      encrypter.Interface
	encryptionKey  string
	masterID       int
	masterSubnet   net.IPNet
	registryDomain string
}

func (e *baseExtension) templateDataTCCP() TemplateData {
	awsRegion := key.Region(e.cluster)

	eniAddress, gateway, subnetSize := calculateNetworkForENI(e.masterSubnet)

	data := TemplateData{
		AWSRegion:           awsRegion,
		AWSConfigSpec:       e.awsConfigSpec,
		IsChinaRegion:       key.IsChinaRegion(awsRegion),
		MasterENIAddress:    eniAddress,
		MasterENIGateway:    gateway,
		MasterENISubnetSize: subnetSize,
		MasterID:            e.masterID,
		RegistryDomain:      e.registryDomain,
	}

	return data
}
func (e *baseExtension) templateDataTCNP() TemplateData {
	awsRegion := key.Region(e.cluster)

	data := TemplateData{
		AWSRegion:      awsRegion,
		AWSConfigSpec:  e.awsConfigSpec,
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

func calculateNetworkForENI(ipNet net.IPNet) (string, string, int) {
	// VPC subnet has reserved first 4 IPs so we need to use the fifth one (counting from zero it is index 4)
	// https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Subnets.html
	eniAddressIP := dupIP(ipNet.IP)
	eniAddressIP[3] += 4

	// gateway is always first available IP in the subnet
	gatewayIP := dupIP(ipNet.IP)
	gatewayIP[3] += 1

	subnetSize, _ := ipNet.Mask.Size()

	return eniAddressIP.String(), gatewayIP.String(), subnetSize
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
