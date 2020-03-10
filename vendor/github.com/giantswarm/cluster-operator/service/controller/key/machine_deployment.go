package key

import (
	"github.com/giantswarm/microerror"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func ToMachineDeployment(v interface{}) (apiv1alpha2.MachineDeployment, error) {
	if v == nil {
		return apiv1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha2.MachineDeployment{}, v)
	}

	p, ok := v.(*apiv1alpha2.MachineDeployment)
	if !ok {
		return apiv1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha2.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
