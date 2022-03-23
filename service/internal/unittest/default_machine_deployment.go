package unittest

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

const (
	DefaultMachineDeploymentID = "al9qy"
)

func DefaultMachineDeployment() infrastructurev1alpha3.AWSMachineDeployment {
	cr := infrastructurev1alpha3.AWSMachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.MachineDeploymentSubnet: "10.100.8.0/24",
			},
			Labels: map[string]string{
				label.Cluster:           DefaultClusterID,
				label.MachineDeployment: DefaultMachineDeploymentID,
				label.OperatorVersion:   "7.3.0",
				label.Release:           "100.0.0",
			},
			Name:      DefaultMachineDeploymentID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha3.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePool{
				Description: "Test node pool for cluster in template rendering unit test.",
				Machine: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolMachine{
					DockerVolumeSizeGB:  100,
					KubeletVolumeSizeGB: 100,
				},
				Scaling: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolScaling{
					Max: 5,
					Min: 3,
				},
			},
			Provider: infrastructurev1alpha3.AWSMachineDeploymentSpecProvider{
				AvailabilityZones: []string{"eu-central-1a", "eu-central-1c"},
				InstanceDistribution: infrastructurev1alpha3.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                0,
					OnDemandPercentageAboveBaseCapacity: to.IntP(100),
				},
				Worker: infrastructurev1alpha3.AWSMachineDeploymentSpecProviderWorker{
					InstanceType:          "m5.2xlarge",
					UseAlikeInstanceTypes: true,
				},
			},
		},
	}

	return cr
}

func MachineDeploymentWithAZs(machineDeployment infrastructurev1alpha3.AWSMachineDeployment, azs []string) infrastructurev1alpha3.AWSMachineDeployment {
	machineDeployment.Spec.Provider.AvailabilityZones = azs

	return machineDeployment
}
