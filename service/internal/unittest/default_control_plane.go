package unittest

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func DefaultContextControlPlane() context.Context {
	ctx := DefaultControllerContext()
	ctx.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID = "snap-1234567890abcdef0"
	cc := controllercontext.NewContext(context.Background(), ctx)
	return cc
}

func DefaultAWSControlPlane() infrastructurev1alpha2.AWSControlPlane {
	cr := infrastructurev1alpha2.AWSControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.ControlPlane:    "a2wax",
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
			},
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: []string{"eu-central-1b"},
			InstanceType:      "m5.xlarge",
		},
	}

	return cr
}

func DefaultAWSControlPlaneWithAZs(azs ...string) infrastructurev1alpha2.AWSControlPlane {
	cp := DefaultAWSControlPlane()
	cp.Spec.AvailabilityZones = azs
	return cp
}

func DefaultG8sControlPlane() infrastructurev1alpha2.G8sControlPlane {
	cr := infrastructurev1alpha2.G8sControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.ControlPlane:    "a2wax",
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
			},
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.G8sControlPlaneSpec{
			Replicas: 1,
		},
	}

	return cr
}
