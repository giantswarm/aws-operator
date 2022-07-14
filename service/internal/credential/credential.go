package credential

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/v12/service/controller/key"
)

const (
	// AWSOperatorArnKey is the key in the Secret under which the ARN for the
	// aws-operator role is held.
	AWSOperatorArnKey = "aws.awsoperator.arn"
)

const (
	DefaultName      = "credential-default"
	DefaultNamespace = "giantswarm"
)

func GetARN(ctx context.Context, k8sClient kubernetes.Interface, cr infrastructurev1alpha3.AWSCluster) (string, error) {
	var err error

	var credential *corev1.Secret
	{
		credentialName := key.CredentialName(cr)
		if credentialName == "" {
			return "", microerror.Mask(credentialNameEmpty)
		}

		credentialNamespace := key.CredentialNamespace(cr)
		if credentialName == "" {
			return "", microerror.Mask(credentialNamespaceEmpty)
		}

		credential, err = k8sClient.CoreV1().Secrets(credentialNamespace).Get(ctx, credentialName, metav1.GetOptions{})
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	arn, err := getARN(credential)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return arn, nil
}

// GetDefaultARN is used only by the bridgezone resource. It should be removed
// when the resource is removed.
func GetDefaultARN(ctx context.Context, k8sClient kubernetes.Interface) (string, error) {
	credential, err := k8sClient.CoreV1().Secrets(DefaultNamespace).Get(ctx, DefaultName, metav1.GetOptions{})
	if err != nil {
		return "", microerror.Mask(err)
	}

	arn, err := getARN(credential)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return arn, nil
}

func getARN(credential *corev1.Secret) (string, error) {
	arn, ok := credential.Data[AWSOperatorArnKey]
	if !ok {
		return "", microerror.Maskf(arnNotFound, AWSOperatorArnKey)
	}

	return string(arn), nil
}
