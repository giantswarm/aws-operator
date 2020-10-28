package key

import (
	"fmt"
	"strconv"

	"github.com/dylanmei/iso8601"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp/template"
)

var (
	// MachineDeploymentLaunchTemplateOverrides is a mapping for instance type
	// overrides. We made these up and can adapt them to our needs. The meaning of
	// the mapping is that e.g. when wanting m4.xlarge but these are unavailable
	// we allow to chose m5.xlarge to fulfil the scaling requirements.
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
		"m4.16xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m5.16xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m4.16xlarge",
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
		"m5.16xlarge": {
			template.LaunchTemplateOverride{
				InstanceType:     "m5.16xlarge",
				WeightedCapacity: 1,
			},
			template.LaunchTemplateOverride{
				InstanceType:     "m4.16xlarge",
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

func MachineDeploymentMetadataV2(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSMetadata]
	if !ok {
		return "optional"
	}
	return result
}

func MachineDeploymentLaunchTemplateName(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return fmt.Sprintf("%s-%s-LaunchTemplate", ClusterID(&cr), MachineDeploymentID(&cr))
}

func MachineDeploymentKubeletVolumeSizeGB(cr infrastructurev1alpha2.AWSMachineDeployment) string {
	return strconv.Itoa(cr.Spec.NodePool.Machine.KubeletVolumeSizeGB)
}

// MachineDeploymentParseMaxBatchSize will try parse the value into valid maxBatchSize
// valid values can be either:
// an integer between 0 < x <= worker count
// a float between 0 < x <= 1
// float value is used as ratio of a total worker count
func MachineDeploymentParseMaxBatchSize(val string, workers int) string {
	// try parse an integer
	integer, err := strconv.Atoi(val)
	if err == nil {
		// check if the value is bigger than zero but lower-or-equal to maximum number of workers
		if integer > 0 && integer <= workers {
			// integer value can be directly used, no need for any adjustment
			return val
		} else {
			// the value is outside of valid bounds, it cannot be used
			return ""
		}
	}
	// try parse float
	ratio, err := strconv.ParseFloat(val, 10)
	if err != nil {
		// not integer or float which means invalid value
		return ""
	}
	// valid value is a decimal representing a percentage
	// anything smaller than 0 or bigger than 1 is not valid
	if ratio > 0 && ratio <= 1.0 {
		// compute the maxBatchSize with the ratio
		maxBatchSize := MachineDeploymentWorkerCountRatio(workers, float32(ratio))

		return maxBatchSize
	}

	return ""
}

// MachineDeploymentPauseTimeIsValid checks if the value is in proper ISO 8601 duration format
func MachineDeploymentPauseTimeIsValid(val string) bool {
	_, err := iso8601.ParseDuration(val)
	if err != nil {
		return false
	}

	return true
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
	return *cr.Spec.Provider.InstanceDistribution.OnDemandPercentageAboveBaseCapacity
}

func MachineDeploymentWorkerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return strconv.Itoa(rounded)
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
