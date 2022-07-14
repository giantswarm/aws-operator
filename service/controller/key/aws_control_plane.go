package key

import (
	"fmt"
	"strconv"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v2/pkg/label"
)

func ControlPlaneAvailabilityZones(cr infrastructurev1alpha3.AWSControlPlane) []string {
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
	return fmt.Sprintf("etcd%d", id)
}

func ControlPlaneID(getter LabelsGetter) string {
	return getter.GetLabels()[label.ControlPlane]
}

func ControlPlaneInstanceType(cr infrastructurev1alpha3.AWSControlPlane) string {
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

func ControlPlaneNodeRole(cr infrastructurev1alpha3.AWSControlPlane) string {
	return fmt.Sprintf("gs-cluster-%s-role-tccpn", ClusterID(&cr))
}

func ControlPlaneRecordSetsRecordValue(id int) string {
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

func ControlPlaneMetadataV2(cr infrastructurev1alpha3.AWSControlPlane) string {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSMetadataV2]
	if !ok {
		return "optional"
	}
	return result
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

func ControlPlaneVolumeIops(cr infrastructurev1alpha3.AWSControlPlane) int {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSEBSVolumeIops]
	if !ok {
		// IOPS will be defaulted if annotaton is not set
		return 0
	}
	o, err := strconv.Atoi(result)
	if err != nil {
		// IOPS will be defaulted when unable to convert properly
		return 0
	}
	return o
}

func ControlPlaneVolumeThroughput(cr infrastructurev1alpha3.AWSControlPlane) int {
	result, ok := cr.ObjectMeta.Annotations[annotation.AWSEBSVolumeThroughput]
	if !ok {
		// Throughput will be defaulted if annotation is not set
		return 0
	}
	o, err := strconv.Atoi(result)
	if err != nil {
		// Throughput will be defaulted when unable to convert to int
		return 0
	}
	return o
}

func ToControlPlane(v interface{}) (infrastructurev1alpha3.AWSControlPlane, error) {
	if v == nil {
		return infrastructurev1alpha3.AWSControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.AWSControlPlane{}, v)
	}

	p, ok := v.(*infrastructurev1alpha3.AWSControlPlane)
	if !ok {
		return infrastructurev1alpha3.AWSControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.AWSControlPlane{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
