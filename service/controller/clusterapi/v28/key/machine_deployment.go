package key

import (
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func WorkerClusterID(cr v1alpha1.MachineDeployment) string {
	return cr.Labels[LabelCluster]
}

// TODO this method has to be properly implemented and renamed eventually.
func StatusAvailabilityZones(cluster v1alpha1.MachineDeployment) []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone {
	return nil
}

func ToMachineDeployment(v interface{}) (v1alpha1.MachineDeployment, error) {
	if v == nil {
		return v1alpha1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.MachineDeployment{}, v)
	}

	p, ok := v.(*v1alpha1.MachineDeployment)
	if !ok {
		return v1alpha1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

func WorkerAvailabilityZones(cr v1alpha1.MachineDeployment) []string {
	return machineDeploymentProviderSpec(cr).Provider.AvailabilityZones
}

func WorkerDockerVolumeSizeGB(cr v1alpha1.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.DockerVolumeSizeGB)
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

func WorkerVersion(cr v1alpha1.MachineDeployment) string {
	return machineDeploymentProviderSpec(cr).NodePool.VersionBundle.Version
}
