package key

import (
	"strconv"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp/template"
)

var (
	MachineDeploymentLaunchTemplateOverrides = map[string][]template.LaunchTemplateOverride{
		"m4.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m4.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m5.xlarge",
				WeightedCapacity: 1,
			},
		},
	}
)

func MachineDeploymentAvailabilityZones(cr infrastructurev1alpha2.AWSMachineDeployment) []string {
	return cr.Spec.Provider.AvailabilityZones
}

func MachineDeploymentDockerVolumeSizeGB(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return strconv.Itoa(cr.Spec.NodePool.Machine.DockerVolumeSizeGB)
}

func MachineDeploymentInstanceType(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return cr.Spec.Provider.Worker.InstanceType
}

func MachineDeploymentKubeletVolumeSizeGB(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return strconv.Itoa(cr.Spec.NodePool.Machine.KubeletVolumeSizeGB)
}

func MachineDeploymentScalingMax(cr infrastructurev1alpha2.AWSMachineDeployment) int {
	return cr.Spec.NodePool.Scaling.Max
}

func MachineDeploymentScalingMin(cr infrastructurev1alpha2.AWSMachineDeployment) int {
	return cr.Spec.NodePool.Scaling.Min
}

func MachineDeploymentSpotInstancePools(cr infrastructurev1alpha2.AWSMachineDeployment, overrides []template.LaunchTemplateOverride) int {
	pools := len(MachineDeploymentAvailabilityZones(cr)) * len(overrides)
	if pools < 1 {
		return 1
	}
	if pools > 20 {
		return 20
	}
	return pools
}

func MachineDeploymentSubnet(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
}

func MachineDeploymentOnDemandBaseCapacity(cr infrastructurev1alpha2.AWSMachineDeployment) int {
	return cr.Spec.Provider.InstanceDistribution.OnDemandBaseCapacity
}

func MachineDeploymentOnDemandPercentageAboveBaseCapacity(cr infrastructurev1alpha2.AWSMachineDeployment) int {
	return cr.Spec.Provider.InstanceDistribution.OnDemandPercentageAboveBaseCapacity
}

func ToMachineDeployment(v interface{}) (infrastructurev1alpha2.AWSMachineDeployment, error) {
	if v == nil {
		return infrastructurev1alpha2.AWSMachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.AWSMachineDeployment{}, v)
	}

	p, ok := v.(*infrastructurev1alpha2.AWSMachineDeployment)
	if !ok {
		return infrastructurev1alpha2.AWSMachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.AWSMachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
