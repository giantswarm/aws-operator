package credential

import (
	"encoding/json"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

const (
	providerKey = "aws"
	roleKey     = "awsoperator"
)

type role struct {
	ARN string `json:"arn"`
}

type roleMap map[string]role

func GetRole(k8sClient kubernetes.Interface, obj interface{}) (*role, error) {
	credential, err := readCredential(k8sClient, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	role, err := getRole(credential)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return role, nil
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

func getRole(credential *v1.Secret) (*role, error) {
	var r roleMap

	data, ok := credential.Data[providerKey]
	if !ok {
		return nil, microerror.Maskf(roleNotFound, providerKey)
	}

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, microerror.Maskf(malformedRole, "%v", err)
	}

	v, ok := r[roleKey]
	if !ok {
		return nil, microerror.Maskf(roleNotFound, roleKey)
	}

	return &v, nil
}
