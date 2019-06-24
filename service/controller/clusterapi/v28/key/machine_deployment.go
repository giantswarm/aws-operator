package key

import (
	"math/bits"
	"net"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func WorkerSubnet(cr v1alpha1.MachineDeployment) string {
	return cr.Labels[label.MachineDeploymentSubnet]
}

func WorkerClusterID(cr v1alpha1.MachineDeployment) string {
	return cr.Labels[label.Cluster]
}

// TODO this method has to be properly implemented and renamed eventually.
func StatusAvailabilityZones(cr v1alpha1.MachineDeployment) []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone {
	var workerSubnet net.IPNet
	{
		s := WorkerSubnet(cr)
		if s == "" {
			//return nil, microerror.Maskf(notFoundError, "MachineDeployment is missing subnet allocation")
			panic("MachineDeployment is missing subnet allocation")
		}

		_, n, err := net.ParseCIDR(s)
		if err != nil {
			//return nil, microerror.Mask(err)
			panic(err)
		}

		workerSubnet = *n
	}

	var azs []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone
	{
		azsSubnets, err := splitNetwork(workerSubnet, uint(len(WorkerAvailabilityZones(cr))))
		if err != nil {
			//return nil, microerror.Mask(err)
			panic(err)
		}

		for i, s := range WorkerAvailabilityZones(cr) {
			subnets, err := splitNetwork(azsSubnets[i], 2)
			if err != nil {
				//return nil, microerror.Mask(err)
				panic(err)
			}

			azs = append(azs, g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				Name: s,
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

// calculateSubnetMask calculates new subnet mask to accommodate n subnets.
func calculateSubnetMask(networkMask net.IPMask, n uint) (net.IPMask, error) {
	if n == 0 {
		return nil, microerror.Maskf(invalidParameterError, "divide by zero")
	}

	// Amount of bits needed to accommodate enough subnets for public and
	// private subnet in each AZ.
	subnetBitsNeeded := bits.Len(n - 1)

	maskOnes, maskBits := networkMask.Size()
	if subnetBitsNeeded > maskBits-maskOnes {
		return nil, microerror.Maskf(invalidParameterError, "no room in network mask %s to accommodate %d subnets", networkMask.String(), n)
	}

	return net.CIDRMask(maskOnes+subnetBitsNeeded, maskBits), nil
}

// splitNetwork returns n subnets from network.
func splitNetwork(network net.IPNet, n uint) ([]net.IPNet, error) {
	mask, err := calculateSubnetMask(network.Mask, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var subnets []net.IPNet
	for i := uint(0); i < n; i++ {
		subnet, err := ipam.Free(network, mask, subnets)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		subnets = append(subnets, subnet)
	}

	return subnets, nil
}
