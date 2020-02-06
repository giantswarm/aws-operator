package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/aws-operator/pkg/label"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultControlPlane() infrastructurev1alpha2.AWSControlPlane {
	cr := infrastructurev1alpha2.AWSControlPlane{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: []string{"eu-central-1b"},
			InstanceType:      "m5.xlarge",
		},
		Status: infrastructurev1alpha2.AWSControlPlaneStatus{
			Status: "",
		},
	}

	return cr
}

func MachineDeploymentWithAZs(machineDeployment infrastructurev1alpha2.AWSMachineDeployment, azs []string) infrastructurev1alpha2.AWSMachineDeployment {
	machineDeployment.Spec.Provider.AvailabilityZones = azs

	return machineDeployment
}
