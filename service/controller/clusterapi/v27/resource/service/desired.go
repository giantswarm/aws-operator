package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

const (
	AnnotationEtcdDomain        = "giantswarm.io/etcd-domain"
	AnnotationPrometheusCluster = "giantswarm.io/prometheus-cluster"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	service := &v1.Service{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      "master",
			Namespace: legacykey.ClusterID(customObject),
			Labels: map[string]string{
				legacykey.LabelApp:           "master",
				legacykey.LabelCluster:       legacykey.ClusterID(customObject),
				legacykey.LabelOrganization:  legacykey.OrganizationID(customObject),
				legacykey.LabelVersionBundle: legacykey.VersionBundleVersion(customObject),
			},
			Annotations: map[string]string{
				AnnotationEtcdDomain:        legacykey.ClusterEtcdEndpoint(customObject),
				AnnotationPrometheusCluster: legacykey.ClusterID(customObject),
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       httpsPort,
					TargetPort: intstr.FromInt(httpsPort),
				},
			},
		},
	}

	return service, nil
}
