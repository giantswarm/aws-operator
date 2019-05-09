package key

import (
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func WorkerInstanceType(machineDeployment v1alpha1.MachineDeployment) string {
	return machineDeploymentProviderSpec(machineDeployment).Provider.Worker.InstanceType
}
