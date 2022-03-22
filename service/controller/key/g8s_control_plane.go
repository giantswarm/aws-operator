package key

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
)

func G8sControlPlaneReplicas(cr infrastructurev1alpha3.G8sControlPlane) int {
	return cr.Spec.Replicas
}
