package unittest

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v12/pkg/label"
	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
)

func DefaultContextControlPlane() context.Context {
	ctx := DefaultControllerContext()
	ctx.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID = "snap-1234567890abcdef0"
	cc := controllercontext.NewContext(context.Background(), ctx)
	return cc
}

func DefaultAWSControlPlane() infrastructurev1alpha3.AWSControlPlane {
	cr := infrastructurev1alpha3.AWSControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.ControlPlane:    "a2wax",
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
			},
			Annotations: map[string]string{},
			Namespace:   metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha3.AWSControlPlaneSpec{
			AvailabilityZones: []string{"eu-central-1b"},
			InstanceType:      "m5.xlarge",
		},
	}

	return cr
}

func DefaultAWSControlPlaneWithAZs(azs ...string) infrastructurev1alpha3.AWSControlPlane {
	cp := DefaultAWSControlPlane()
	cp.Spec.AvailabilityZones = azs
	return cp
}

func DefaultG8sControlPlane() infrastructurev1alpha3.G8sControlPlane {
	cr := infrastructurev1alpha3.G8sControlPlane{
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
		Spec: infrastructurev1alpha3.G8sControlPlaneSpec{
			Replicas: 1,
		},
	}

	return cr
}
