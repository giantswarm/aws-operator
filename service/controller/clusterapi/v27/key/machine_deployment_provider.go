package key

import (
	"encoding/json"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func mustG8sMachineDeploymentSpecFromCMASpec(cmaSpec cmav1alpha1.ProviderSpec) g8sv1alpha1.AWSMachineDeploymentSpec {
	if cmaSpec.Value == nil {
		panic("provider spec value must not be empty")
	}

	var g8sSpec g8sv1alpha1.AWSMachineDeploymentSpec
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

func mustG8sMachineDeploymentStatusFromCMAStatus(cmaStatus *runtime.RawExtension) g8sv1alpha1.AWSMachineDeploymentStatus {
	if cmaStatus == nil {
		panic("provider status value must not be empty")
	}

	var g8sStatus g8sv1alpha1.AWSMachineDeploymentStatus
	{
		if len(cmaStatus.Raw) == 0 {
			return g8sStatus
		}

		err := json.Unmarshal(cmaStatus.Raw, &g8sStatus)
		if err != nil {
			panic(err)
		}
	}

	return g8sStatus
}

func machineDeploymentProviderSpec(machineDeployment cmav1alpha1.MachineDeployment) g8sv1alpha1.AWSMachineDeploymentSpec {
	return mustG8sMachineDeploymentSpecFromCMASpec(machineDeployment.Spec.Template.Spec.ProviderSpec)
}

func machineDeploymentProviderStatus(machineDeployment cmav1alpha1.MachineDeployment) g8sv1alpha1.AWSMachineDeploymentStatus {
	return mustG8sMachineDeploymentStatusFromCMAStatus(machineDeployment.Status.)
}
