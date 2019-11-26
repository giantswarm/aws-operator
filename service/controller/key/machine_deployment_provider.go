package key

import (
	"encoding/json"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
)

func machineDeploymentProviderSpec(machineDeployment infrastructurev1alpha2.AWSMachineDeployment) infrastructurev1alpha2.AWSMachineDeploymentSpec {
	return mustG8sMachineDeploymentSpecFromCMAMachineDeploymentSpec(machineDeployment.Spec.Template.Spec.ProviderSpec)
}

func mustG8sMachineDeploymentSpecFromCMAMachineDeploymentSpec(cmaSpec infrastructurev1alpha2.ProviderSpec) infrastructurev1alpha2.AWSMachineDeploymentSpec {
	if cmaSpec.Value == nil {
		panic("provider spec value must not be empty")
	}

	var g8sSpec infrastructurev1alpha2.AWSMachineDeploymentSpec
	{
		if len(cmaSpec.Value.Raw) == 0 {
			return g8sSpec
		}

		err := json.Unmarshal(cmaSpec.Value.Raw, &g8sSpec)
		if err != nil {
			panic(err)
		}
	}

	return g8sSpec
}
