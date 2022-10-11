package service

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/aws-operator/v14/pkg/label"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

const (
	AnnotationEtcdDomain        = "giantswarm.io/etcd-domain"
	AnnotationPrometheusCluster = "giantswarm.io/prometheus-cluster"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "master",
			Namespace: key.ClusterID(&cr),
			Labels: map[string]string{
				label.App:          "master",
				label.Cluster:      key.ClusterID(&cr),
				label.Organization: key.OrganizationID(&cr),
			},
			Annotations: map[string]string{
				AnnotationEtcdDomain:        key.ClusterEtcdEndpointWithPort(cr),
				AnnotationPrometheusCluster: key.ClusterID(&cr),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       httpsPort,
					TargetPort: intstr.FromInt(httpsPort),
				},
			},
		},
	}

	return service, nil
}
