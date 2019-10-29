package unittest

import (
	"encoding/json"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

func DefaultMachineDeployment() cmav1alpha1.MachineDeployment {
	cr := cmav1alpha1.MachineDeployment{
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

	spec := g8sv1alpha1.AWSMachineDeploymentSpec{
		NodePool: g8sv1alpha1.AWSMachineDeploymentSpecNodePool{
			Description: "Test node pool for cluster in template rendering unit test.",
			Machine: g8sv1alpha1.AWSMachineDeploymentSpecNodePoolMachine{
				DockerVolumeSizeGB:  100,
				KubeletVolumeSizeGB: 100,
			},
			Scaling: g8sv1alpha1.AWSMachineDeploymentSpecNodePoolScaling{
				Max: 5,
				Min: 3,
			},
		},
		Provider: g8sv1alpha1.AWSMachineDeploymentSpecProvider{
			AvailabilityZones: []string{"eu-central-1a", "eu-central-1c"},
			Worker: g8sv1alpha1.AWSMachineDeploymentSpecProviderWorker{
				InstanceType: "m5.2xlarge",
			},
		},
	}

	return mustCMAMachineDeploymentWithG8sProviderSpec(cr, spec)
}

func MachineDeploymentWithAZs(machineDeployment cmav1alpha1.MachineDeployment, azs []string) cmav1alpha1.MachineDeployment {
	spec := mustG8sMachineDeploymentSpecFromCMAMachineDeploymentSpec(machineDeployment.Spec.Template.Spec.ProviderSpec)

	spec.Provider.AvailabilityZones = azs

	return mustCMAMachineDeploymentWithG8sProviderSpec(machineDeployment, spec)
}

func mustG8sMachineDeploymentSpecFromCMAMachineDeploymentSpec(cmaSpec cmav1alpha1.ProviderSpec) g8sv1alpha1.AWSMachineDeploymentSpec {
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

func mustCMAMachineDeploymentWithG8sProviderSpec(cr cmav1alpha1.MachineDeployment, providerExtension g8sv1alpha1.AWSMachineDeploymentSpec) cmav1alpha1.MachineDeployment {
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
