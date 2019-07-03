package key

import (
	"net"
	"sort"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/network"
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

// TODO this method has to be properly implemented and renamed eventually.
func StatusAvailabilityZones(cr v1alpha1.MachineDeployment) ([]g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone, error) {
	var workerSubnet net.IPNet
	{
		s := WorkerSubnet(cr)
		if s == "" {
			return nil, microerror.Maskf(notFoundError, "MachineDeployment is missing subnet allocation")
		}

		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		workerSubnet = *n
	}

	var azs []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone
	{
		workerAZs := SortedWorkerAvailabilityZones(cr)

		if len(workerAZs) > MaxNumberOfAZs {
			return nil, microerror.Maskf(invalidParameterError, "too many availability zones defined: %d, max: %d", len(workerAZs), MaxNumberOfAZs)
		}

		// In order to have room for dynamically changing number of AZs we
		// reserve always $MaxNumberOfAZs number of subnets from cluster
		// network.
		azsSubnets, err := network.Split(workerSubnet, MaxNumberOfAZs)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// For temporarily testing CloudFormation templates & adapters for
		// dynamically changing AZs, we dedicate subnets from above to AZs in
		// alphabetical order. This is only temporary solution to further test
		// and develop changes for CloudFormation stack changes.
		//
		// In the final setting this will be changed so that whenever new AZ
		// needs subnet, it's properly allocated by first gathering reserved
		// subnets from corresponding cluster VPC.
		azSubnets := make(map[string]net.IPNet)
		{
			var subnetList [MaxNumberOfAZs]net.IPNet

			for i, s := range azsSubnets {
				subnetList[i] = s
			}

			// Take last letter of AZ (i.e. a, b, c or d for now) and compute
			// list index from it and dedicate corresponding subnet for it.
			// This way we make subnet allocation for AZs deterministic even
			// when selected AZs are not consecutive.
			for _, az := range workerAZs {
				zone := az[len(az)-1]
				idx := zone - 'a'
				azSubnets[az] = subnetList[idx]
			}
		}

		// Finally split each AZ specific subnet into public and private part.
		for _, az := range workerAZs {
			subnets, err := network.Split(azSubnets[az], 2)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			azs = append(azs, g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				Name: az,
				Subnet: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
					Private: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
						CIDR: subnets[0].String(),
					},
					Public: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
						CIDR: subnets[1].String(),
					},
				},
			})
		}
	}

	return azs, nil
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
