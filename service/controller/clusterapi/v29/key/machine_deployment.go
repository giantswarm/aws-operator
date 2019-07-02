package key

import (
	"net"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/network"
)

func WorkerClusterID(cr v1alpha1.MachineDeployment) string {
	return cr.Labels[label.Cluster]
}

func WorkerSubnet(cr v1alpha1.MachineDeployment) string {
	return cr.Annotations[annotation.MachineDeploymentSubnet]
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
		workerAZs := WorkerAvailabilityZones(cr)

		azsSubnets, err := network.Split(workerSubnet, uint(len(workerAZs)))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for i, s := range workerAZs {
			subnets, err := network.Split(azsSubnets[i], 2)
			if err != nil {
				return nil, microerror.Mask(err)
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
