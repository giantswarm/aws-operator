package key

import (
	"fmt"

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

func ControlPlaneENIName(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-eni", ClusterID(getter), masterID)
}

func ControlPlaneLaunchTemplateName(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-launch-template", ClusterID(getter), masterID)
}

func ControlPlaneVolumeNameEtcd(getter LabelsGetter, masterID int) string {
	return fmt.Sprintf("%s-master%d-etcd", ClusterID(getter), masterID)
}
