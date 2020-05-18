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

func ControlPlaneASGResourceName(getter LabelsGetter, id int) string {
	if id == 0 || id == 1 {
		return "ControlPlaneNodeAutoScalingGroup"
	}

	return fmt.Sprintf("ControlPlaneNodeAutoScalingGroup%d", id)
}

func ControlPlaneENIName(getter LabelsGetter, id int) string {
	return fmt.Sprintf("%s-master%d-eni", ClusterID(getter), id)
}

func ControlPlaneENIResourceName(id int) string {
	if id == 0 || id == 1 {
		return "MasterEni"
	}

	return fmt.Sprintf("MasterEni%d", id)
}

func ControlPlaneEtcdNodeName(id int) string {
	if id == 0 {
		return "etcd1"
	}
	return fmt.Sprintf("etcd%d", id)
}

func ControlPlaneID(getter LabelsGetter) string {
	return getter.GetLabels()[label.ControlPlane]
}

func ControlPlaneInstanceType(cr infrastructurev1alpha2.AWSControlPlane) string {
	return cr.Spec.InstanceType
}

func ControlPlaneLaunchTemplateName(getter LabelsGetter, id int) string {
	return fmt.Sprintf("%s-master%d-launch-template", ClusterID(getter), id)
}

func ControlPlaneLaunchTemplateResourceName(getter LabelsGetter, id int) string {
	if id == 0 || id == 1 {
		return "ControlPlaneNodeLaunchTemplate"
	}

	return fmt.Sprintf("ControlPlaneNodeLaunchTemplate%d", id)
}

func ControlPlaneRecordSetsRecordValue(id int) string {
	if id == 0 {
		return "etcd1"
	}
	return fmt.Sprintf("etcd%d", id)
}

func ControlPlaneRecordSetsResourceName(id int) string {
	if id == 0 || id == 1 {
		return "ControlPlaneRecordSet"
	}

	return fmt.Sprintf("ControlPlaneRecordSet%d", id)
}

func ControlPlaneVolumeName(getter LabelsGetter, id int) string {
	return fmt.Sprintf("%s-master%d-etcd", ClusterID(getter), id)
}

func ControlPlaneVolumeResourceName(id int) string {
	if id == 0 || id == 1 {
		return "EtcdVolume"
	}

	return fmt.Sprintf("EtcdVolume%d", id)
}

func ControlPlaneVolumeSnapshotID(snapshot string, master int) string {
	if master == 0 || master == 1 {
		// Master ID 0 does only exist in single master setups. Master ID 1 does
		// only exist in HA Masters setups. In either setup it does only work to
		// provide a Snapshot ID for one of the running masters, of which other
		// masters replicate in a HA Masters setup. For backward compatability we
		// maintain the Snapshot ID of Tenant Clusters upgrading to this version so
		// that there is an automated migration path.
		return snapshot
	}

	return ""
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
