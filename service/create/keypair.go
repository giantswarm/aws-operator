package create

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

type keyPairProvider interface {
	pubKeyContent() ([]byte, error)
}

type fsKeyPairProvider struct {
	pubKeyFile string
}

func newFsKeyPairProvider(pubKeyFile string) *fsKeyPairProvider {
	return &fsKeyPairProvider{
		pubKeyFile: pubKeyFile,
	}
}

func (f *fsKeyPairProvider) pubKeyContent() ([]byte, error) {
	return ioutil.ReadFile(f.pubKeyFile)
}

type keyPairInput struct {
	ec2Client   *ec2.EC2
	clusterName string
	provider    keyPairProvider
}

func (s *Service) keyPair(input keyPairInput) (string, error) {
	pkc, err := input.provider.pubKeyContent()
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	keyPair, err := input.ec2Client.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String(input.clusterName),
		PublicKeyMaterial: pkc,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == awsclient.KeyPairDuplicate {
			s.Logger.Log("info", fmt.Sprintf("keypair '%s' exists, reusing", input.clusterName))
			return input.clusterName, nil
		} else {
			return "", microerror.MaskAny(err)
		}
	}

	if keyPair == nil || keyPair.KeyName == nil {
		return "", fmt.Errorf("Couln't create and find the keypair '%s'", input.clusterName)
	}

	return *keyPair.KeyName, nil
}
