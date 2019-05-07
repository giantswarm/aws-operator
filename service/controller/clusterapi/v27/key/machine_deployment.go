package key

import (
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func WorkerInstanceType(machineDeployment v1alpha1.MachineDeployment) string {
	var instanceType string

	if len(machineDeployment.Spec.AWS.Workers) > 0 {
		instanceType = machineDeployment.Spec.AWS.Workers[0].InstanceType

	}

	return instanceType
}
