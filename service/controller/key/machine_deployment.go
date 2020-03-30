package key

import (
	"fmt"
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
		"m4.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m4.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m5.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"m4.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m4.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m5.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"m5.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m5.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m4.xlarge",
				WeightedCapacity: 1,
			},
		},
		"m5.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m5.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m4.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"m5.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m5.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m4.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"c4.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c4.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c5.xlarge",
				WeightedCapacity: 1,
			},
		},
		"c4.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c4.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c5.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"c4.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c4.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c5.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"c4.8xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c4.8xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c5.8xlarge",
				WeightedCapacity: 1,
			},
		},
		"c5.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c5.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c4.xlarge",
				WeightedCapacity: 1,
			},
		},
		"c5.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c5.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c4.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"c5.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c5.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c4.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"c5.8xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "c5.8xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "c4.8xlarge",
				WeightedCapacity: 1,
			},
		},
		"r4.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r4.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r5.xlarge",
				WeightedCapacity: 1,
			},
		},
		"r4.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r4.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r5.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"r4.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r4.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r5.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"r4.8xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r4.8xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r5.8xlarge",
				WeightedCapacity: 1,
			},
		},
		"r5.xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r5.xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r4.xlarge",
				WeightedCapacity: 1,
			},
		},
		"r5.2xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r5.2xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r4.2xlarge",
				WeightedCapacity: 1,
			},
		},
		"r5.4xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r5.4xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r4.4xlarge",
				WeightedCapacity: 1,
			},
		},
		"r5.8xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "r5.8xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "r4.8xlarge",
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

func MachineDeploymentLaunchTemplateName(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return fmt.Sprintf("%s-%s-LaunchTemplate", ClusterID(&cr), MachineDeploymentID(&cr))
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

// MachineDeploymentSpotInstancePools ensures that the number of spot instance pools
// we submit to AWS is within AWS limits of more than 0 and less than 21 pools.
func MachineDeploymentSpotInstancePools(overrides []template.LaunchTemplateOverride) int {
	pools := len(overrides)
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
