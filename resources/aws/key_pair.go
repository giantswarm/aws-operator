package aws

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

type KeyPairProvider interface {
	pubKeyContent() ([]byte, error)
}

type FSKeyPairProvider struct {
	pubKeyFile string
}

func NewFSKeyPairProvider(pubKeyFile string) *FSKeyPairProvider {
	return &FSKeyPairProvider{
		pubKeyFile: pubKeyFile,
	}
}

func (f *FSKeyPairProvider) pubKeyContent() ([]byte, error) {
	return ioutil.ReadFile(f.pubKeyFile)
}

type KeyPair struct {
	ClusterName string
	Provider    KeyPairProvider
	AWSEntity
}

func (k *KeyPair) CreateIfNotExists() (bool, error) {
	err := k.CreateOrFail()
	if err != nil {
		if strings.Contains(err.Error(), awsclient.KeyPairDuplicate) {
			return false, nil
		}
		return false, microerror.Mask(err)
	}
	return true, nil
}

func (k *KeyPair) CreateOrFail() error {
	pkc, err := k.Provider.pubKeyContent()
	if err != nil {
		return microerror.Mask(err)
	}

	keyPair, err := k.Clients.EC2.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String(k.ClusterName),
		PublicKeyMaterial: pkc,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	if keyPair == nil || keyPair.KeyName == nil {
		return fmt.Errorf("Couln't create and find the keypair '%s'", k.ClusterName)
	}

	return nil
}

func (k *KeyPair) Delete() error {
	if _, err := k.Clients.EC2.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(k.ClusterName),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
