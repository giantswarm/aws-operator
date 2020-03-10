package key

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
)

func ToG8sControlPlane(v interface{}) (infrastructurev1alpha2.G8sControlPlane, error) {
	if v == nil {
		return infrastructurev1alpha2.G8sControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.G8sControlPlane{}, v)
	}

	p, ok := v.(*infrastructurev1alpha2.G8sControlPlane)
	if !ok {
		return infrastructurev1alpha2.G8sControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.G8sControlPlane{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
