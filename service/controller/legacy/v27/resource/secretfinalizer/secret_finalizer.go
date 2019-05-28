package secretfinalizer

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func newSecretFinalizer(secret *corev1.Secret) string {
	return fmt.Sprintf("aws-operator.giantswarm.io/%s", secret.Name)
}
