package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/aws-operator/service/controller/v14patch1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	service := &v1.Service{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      "master",
			Namespace: key.ClusterID(customObject),
			Labels: map[string]string{
				key.LabelApp:           "master",
				key.LegacyLabelCluster: key.ClusterID(customObject),
				key.LabelCustomer:      key.CustomerID(customObject),
				key.LabelCluster:       key.ClusterID(customObject),
				key.LabelOrganization:  key.CustomerID(customObject),
				key.LabelVersionBundle: key.VersionBundleVersion(customObject),
			},
			Annotations: map[string]string{
				key.AnnotationEtcdDomain:        key.ClusterEtcdDomain(customObject),
				key.AnnotationPrometheusCluster: key.ClusterID(customObject),
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
