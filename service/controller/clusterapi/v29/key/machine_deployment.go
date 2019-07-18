package key

import (
	"sort"
	"strconv"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
)

// As a first version of Node Pools feature, the maximum number of distinct
// Availability Zones is restricted to four due to current IPAM architecture &
// implementation.
const MaxNumberOfAZs = 4

var AZLetters []byte

func init() {
	alphabets := "abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < MaxNumberOfAZs && i < len(alphabets); i++ {
		AZLetters = append(AZLetters, alphabets[i])
	}
}

func SortedWorkerAvailabilityZones(cr v1alpha1.MachineDeployment) []string {
	azs := WorkerAvailabilityZones(cr)

	// No need to do deep copy for azs slice since above key function
	// deserializes information from provider extension template that is JSON
	// in CR object.

	sort.Slice(azs, func(i, j int) bool {
		return azs[i] < azs[j]
	})

	return azs
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

func WorkerClusterID(cr v1alpha1.MachineDeployment) string {
	return cr.Labels[label.Cluster]
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

func WorkerSubnet(cr v1alpha1.MachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
}
