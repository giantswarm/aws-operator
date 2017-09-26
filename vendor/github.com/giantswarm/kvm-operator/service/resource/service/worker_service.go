package service

import (
	"github.com/giantswarm/kvmtpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func newWorkerService(customObject kvmtpr.CustomObject) *apiv1.Service {
	service := &apiv1.Service{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.WorkerID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
				"app":      key.WorkerID,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Port:       int32(30010),
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromInt(30010),
				},
				{
					Name:       "https",
					Port:       int32(30011),
					Protocol:   apiv1.ProtocolTCP,
					TargetPort: intstr.FromInt(30011),
				},
			},
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service
}
