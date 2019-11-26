package unittest

import (
	"encoding/json"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

func DefaultMachineDeployment() infrastructurev1alpha2.AWSMachineDeployment {
	cr := infrastructurev1alpha2.AWSMachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				annotation.MachineDeploymentSubnet: "10.100.8.0/24",
			},
			Labels: map[string]string{
				label.Cluster:           "8y5ck",
				label.MachineDeployment: "al9qy",
				label.OperatorVersion:   "7.3.0",
			},
		},
	}

	spec := infrastructurev1alpha2.AWSMachineDeploymentSpec{
		NodePool: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePool{
			Description: "Test node pool for cluster in template rendering unit test.",
			Machine: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolMachine{
				DockerVolumeSizeGB:  100,
				KubeletVolumeSizeGB: 100,
			},
			Scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Max: 5,
				Min: 3,
			},
		},
		Provider: infrastructurev1alpha2.AWSMachineDeploymentSpecProvider{
			AvailabilityZones: []string{"eu-central-1a", "eu-central-1c"},
			Worker: infrastructurev1alpha2.AWSMachineDeploymentSpecProviderWorker{
				InstanceType: "m5.2xlarge",
			},
		},
	}

	return mustCMAMachineDeploymentWithG8sProviderSpec(cr, spec)
}

func MachineDeploymentWithAZs(machineDeployment infrastructurev1alpha2.AWSMachineDeployment, azs []string) infrastructurev1alpha2.AWSMachineDeployment {
	spec := mustG8sMachineDeploymentSpecFromCMAMachineDeploymentSpec(machineDeployment.Spec.Template.Spec.ProviderSpec)

	spec.Provider.AvailabilityZones = azs

	return mustCMAMachineDeploymentWithG8sProviderSpec(machineDeployment, spec)
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

func mustCMAMachineDeploymentWithG8sProviderSpec(cr infrastructurev1alpha2.AWSMachineDeployment, providerExtension infrastructurev1alpha2.AWSMachineDeploymentSpec) infrastructurev1alpha2.AWSMachineDeployment {
	var err error

	if cr.Spec.Template.Spec.ProviderSpec.Value == nil {
		cr.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{}
	}

	cr.Spec.Template.Spec.ProviderSpec.Value.Raw, err = json.Marshal(&providerExtension)
	if err != nil {
		panic(err)
	}

	return cr
}
