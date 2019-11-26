package key

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/annotation"
)

func MachineDeploymentAvailabilityZones(cr infrastructurev1alpha2.MachineDeployment) []string {
	return machineDeploymentProviderSpec(cr).Provider.AvailabilityZones
}

func MachineDeploymentDockerVolumeSizeGB(cr infrastructurev1alpha2.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.DockerVolumeSizeGB)
}

func MachineDeploymentInstanceType(cr infrastructurev1alpha2.MachineDeployment) string {
	return machineDeploymentProviderSpec(cr).Provider.Worker.InstanceType
}

func MachineDeploymentKubeletVolumeSizeGB(cr infrastructurev1alpha2.MachineDeployment) string {
	return strconv.Itoa(machineDeploymentProviderSpec(cr).NodePool.Machine.KubeletVolumeSizeGB)
}

func MachineDeploymentScalingMax(cr infrastructurev1alpha2.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Max
}

func MachineDeploymentScalingMin(cr infrastructurev1alpha2.MachineDeployment) int {
	return machineDeploymentProviderSpec(cr).NodePool.Scaling.Min
}

func MachineDeploymentSubnet(cr infrastructurev1alpha2.MachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
}

func ToMachineDeployment(v interface{}) (infrastructurev1alpha2.MachineDeployment, error) {
	if v == nil {
		return infrastructurev1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.MachineDeployment{}, v)
	}

	p, ok := v.(*infrastructurev1alpha2.MachineDeployment)
	if !ok {
		return infrastructurev1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
