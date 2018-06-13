package credential

import (
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

const (
	// awsOperatorArnKey is the key in the Secret under which the ARN for the aws-operator role is held.
	awsOperatorArnKey = "aws.awsoperator.arn"
)

func GetARN(k8sClient kubernetes.Interface, obj interface{}) (string, error) {
	credential, err := readCredential(k8sClient, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	arn, err := getARN(credential)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return arn, nil
}

func getARN(credential *v1.Secret) (string, error) {
	arn, ok := credential.Data[awsOperatorArnKey]
	if !ok {
		return "", microerror.Maskf(arnNotFound, awsOperatorArnKey)
	}

	return string(arn), nil
}

func readCredential(k8sClient kubernetes.Interface, obj interface{}) (*v1.Secret, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	credentialName := key.CredentialName(customObject)
	credentialNamespace := key.CredentialNamespace(customObject)

	credential, err := k8sClient.CoreV1().Secrets(credentialNamespace).Get(credentialName, apismetav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return credential, nil
}
