package key

import (
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// TODO this method has to be properly implemented and renamed eventually.
func StatusAvailabilityZones(cluster v1alpha1.MachineDeployment) []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone {
	return nil
}

func WorkerAvailabilityZones(cr v1alpha1.MachineDeployment) []string {
	return machineDeploymentProviderSpec(cr).Provider.AvailabilityZones
}

func WorkerInstanceType(cr v1alpha1.MachineDeployment) string {
	return machineDeploymentProviderSpec(cr).Provider.Worker.InstanceType
}

func WorkerScalingMax(cr v1alpha1.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Max
}

func WorkerScalingMin(cr v1alpha1.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Min
}

func WorkerDockerVolumeSizeGB(cr v1alpha1.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.DockerVolumeSizeGB)
}
