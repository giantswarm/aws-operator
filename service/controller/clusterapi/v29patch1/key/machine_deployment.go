package key

import (
	"strconv"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
)

func MachineDeploymentAvailabilityZones(cr v1alpha1.MachineDeployment) []string {
	return machineDeploymentProviderSpec(cr).Provider.AvailabilityZones
}

func MachineDeploymentDockerVolumeSizeGB(cr v1alpha1.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.DockerVolumeSizeGB)
}

func MachineDeploymentInstanceType(cr v1alpha1.MachineDeployment) string {
	return machineDeploymentProviderSpec(cr).Provider.Worker.InstanceType
}

func MachineDeploymentKubeletVolumeSizeGB(cr v1alpha1.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.KubeletVolumeSizeGB)
}

func MachineDeploymentScalingMax(cr v1alpha1.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Max
}

func MachineDeploymentScalingMin(cr v1alpha1.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Min
}

func MachineDeploymentSubnet(cr v1alpha1.MachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
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
