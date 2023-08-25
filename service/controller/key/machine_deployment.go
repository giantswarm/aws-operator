package key

import (
	"fmt"
	"strconv"

	"github.com/dylanmei/iso8601"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v14/service/controller/resource/tcnp/template"
)

func MachineDeploymentAvailabilityZones(cr infrastructurev1alpha3.AWSMachineDeployment) []string {
	return cr.Spec.Provider.AvailabilityZones
}

func MachineDeploymentDockerVolumeSizeGB(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return strconv.Itoa(cr.Spec.NodePool.Machine.DockerVolumeSizeGB)
}

func MachineDeploymentContainerdVolumeSizeGB(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	defaultValue := MachineDeploymentDockerVolumeSizeGB(cr)

	//If there is no tag, default to docker volume size
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSContainerdVolumeSize]
	if !ok {
		return defaultValue
	}
	//If the value of the tag is not a number, default to docker volume size
	_, error := strconv.Atoi(result)
	if error == nil {
		return result
	} else {
		return defaultValue
	}
}

func MachineDeploymentFlatcarReleaseVersion(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	result, ok := cr.ObjectMeta.Annotations[annotation.FlatcarReleaseVersion]

	if !ok {
		return ""
	}

	return result
}

func MachineDeploymentLoggingVolumeSizeGB(cr infrastructurev1alpha3.AWSMachineDeployment) int {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSLoggingVolumeSize]
	//If there is no tag, default to 15Gb
	if !ok {
		return 15
	}
	value, error := strconv.Atoi(result)
	//If the content of the tag is not a number, default to 15Gb
	if error != nil {
		return 15
	}
	return value
}

func MachineDeploymentInstanceType(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return cr.Spec.Provider.Worker.InstanceType
}

func MachineDeploymentMetadataV2(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSMetadataV2]
	if !ok {
		return "optional"
	}
	return result
}

func MachineDeploymentLaunchTemplateName(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return fmt.Sprintf("%s-%s-LaunchTemplate", ClusterID(&cr), MachineDeploymentID(&cr))
}

func MachineDeploymentKubeletVolumeSizeGB(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return strconv.Itoa(cr.Spec.NodePool.Machine.KubeletVolumeSizeGB)
}

// MachineDeploymentParseMaxBatchSize will try parse the value into valid maxBatchSize
// valid values can be either:
// an integer between 0 < x
// a float between 0 < x <= 1
// float value is used as ratio of a total worker count
func MachineDeploymentParseMaxBatchSize(val string, workers int) string {
	// try parse an integer
	integer, err := strconv.Atoi(val)
	if err == nil {
		// check if the value is bigger than zero
		if integer > 0 {
			// integer value can be directly used, no need for any adjustment
			return val
		} else {
			// the value is outside of valid bounds, it cannot be used
			return ""
		}
	}
	// try parse float
	ratio, err := strconv.ParseFloat(val, 32)
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

// MachineDeploymentMinInstanceInServiceFromMaxBatchSize will calculate the minInstanceInService
// value for ASG, the value is calculated by subtracting  maxBatchSize from the worker count
func MachineDeploymentMinInstanceInServiceFromMaxBatchSize(maxBatchSize string, workers int) (string, error) {
	v, err := strconv.Atoi(maxBatchSize)
	if err != nil {
		return "", microerror.Mask(err)
	}

	o := workers - v
	if o < 0 {
		o = 0
	}

	return strconv.Itoa(o), nil
}

// MachineDeploymentPauseTimeIsValid checks if the value is in proper ISO 8601 duration format
// and ensure that the duration is not bigger than 1 Hour
func MachineDeploymentPauseTimeIsValid(val string) bool {
	d, err := iso8601.ParseDuration(val)
	if err != nil {
		return false
	}

	// AWS limits the duration to 1 hour
	if d.Hours() > 1.0 {
		return false
	}

	return true
}

func MachineDeploymentScalingMax(cr infrastructurev1alpha3.AWSMachineDeployment) int {
	return cr.Spec.NodePool.Scaling.Max
}

func MachineDeploymentScalingMin(cr infrastructurev1alpha3.AWSMachineDeployment) int {
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

func MachineDeploymentSubnet(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
}

func MachineDeploymentOnDemandBaseCapacity(cr infrastructurev1alpha3.AWSMachineDeployment) int {
	return cr.Spec.Provider.InstanceDistribution.OnDemandBaseCapacity
}

func MachineDeploymentOnDemandPercentageAboveBaseCapacity(cr infrastructurev1alpha3.AWSMachineDeployment) int {
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

func MachineDeploymentNodeRole(cr infrastructurev1alpha3.AWSMachineDeployment) string {
	return fmt.Sprintf("gs-cluster-%s-role-%s", ClusterID(&cr), MachineDeploymentID(&cr))
}

func ToMachineDeployment(v interface{}) (infrastructurev1alpha3.AWSMachineDeployment, error) {
	if v == nil {
		return infrastructurev1alpha3.AWSMachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.AWSMachineDeployment{}, v)
	}

	p, ok := v.(*infrastructurev1alpha3.AWSMachineDeployment)
	if !ok {
		return infrastructurev1alpha3.AWSMachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.AWSMachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
