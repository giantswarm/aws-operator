package key

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
)

func G8sControlPlaneReplicas(cr infrastructurev1alpha2.G8sControlPlane) int {
	return cr.Spec.Replicas
}
