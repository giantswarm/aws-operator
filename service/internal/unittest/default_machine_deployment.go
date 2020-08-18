package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

const (
	DefaultMachineDeploymentID = "al9qy"
)

func DefaultMachineDeployment() infrastructurev1alpha2.AWSMachineDeployment {
	cr := infrastructurev1alpha2.AWSMachineDeployment{
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
		Spec: infrastructurev1alpha2.AWSMachineDeploymentSpec{
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
				InstanceDistribution: infrastructurev1alpha2.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                0,
					OnDemandPercentageAboveBaseCapacity: to.IntP(100),
				},
				Worker: infrastructurev1alpha2.AWSMachineDeploymentSpecProviderWorker{
					InstanceType:          "m5.2xlarge",
					UseAlikeInstanceTypes: true,
				},
			},
		},
	}

	return cr
}

func MachineDeploymentWithAZs(machineDeployment infrastructurev1alpha2.AWSMachineDeployment, azs []string) infrastructurev1alpha2.AWSMachineDeployment {
	machineDeployment.Spec.Provider.AvailabilityZones = azs

	return machineDeployment
}
