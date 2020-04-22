package key

import (
	"fmt"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func ControlPlaneAvailabilityZones(cr infrastructurev1alpha2.AWSControlPlane) []string {
	return cr.Spec.AvailabilityZones
}

func ControlPlaneID(getter LabelsGetter) string {
	return getter.GetLabels()[label.ControlPlane]
}

func ControlPlaneInstanceType(cr infrastructurev1alpha2.AWSControlPlane) string {
	return cr.Spec.InstanceType
}

func ToControlPlane(v interface{}) (infrastructurev1alpha2.AWSControlPlane, error) {
	if v == nil {
		return infrastructurev1alpha2.AWSControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.AWSControlPlane{}, v)
	}

	p, ok := v.(*infrastructurev1alpha2.AWSControlPlane)
	if !ok {
		return infrastructurev1alpha2.AWSControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha2.AWSControlPlane{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

func ControlPlaneENIIpAddress(ipNet net.IPNet) string {
	// VPC subnet has reserved first 4 IPs so we need to use the fifth one (counting from zero it is index 4)
	// https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Subnets.html
	eniAddressIP := dupIP(ipNet.IP)
	eniAddressIP.To4()
	eniAddressIP[3] += 4

	return eniAddressIP.String()
}

func ControlPlaneENIGateway(ipNet net.IPNet) string {
	// https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Subnets.html
	gatewayAddressIP := dupIP(ipNet.IP)
	gatewayAddressIP.To4()
	gatewayAddressIP[3] += 1

	return gatewayAddressIP.String()
}

func ControlPlaneENIName(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-eni", ClusterID(getter), masterID)
}

func ControlPlaneENISubnetSize(ipNet net.IPNet) int {
	subnetSize, _ := ipNet.Mask.Size()

	return subnetSize
}

func ControlPlaneLaunchTemplateName(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-launch-template", ClusterID(getter), masterID)
}

func ControlPlaneVolumeNameEtcd(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-etcd", ClusterID(getter), masterID)
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
