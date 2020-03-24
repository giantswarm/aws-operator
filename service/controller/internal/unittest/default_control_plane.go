package unittest

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultContextControlPlane() context.Context {
	ctx := DefaultControllerContext()
	ctx.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID = "snap-1234567890abcdef0"
	cc := controllercontext.NewContext(context.Background(), ctx)
	return cc
}

func DefaultControlPlane() infrastructurev1alpha2.AWSControlPlane {
	cr := infrastructurev1alpha2.AWSControlPlane{
		ObjectMeta: v1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: []string{"eu-central-1b"},
			InstanceType:      "m5.xlarge",
		},
	}

	return cr
}
