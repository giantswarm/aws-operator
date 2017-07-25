package create

import (
	"github.com/giantswarm/clustertpr"
	microerror "github.com/giantswarm/microkit/error"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func (s *Service) createClusterNamespace(cluster clustertpr.Spec) error {
	namespace := v1.Namespace{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: cluster.Cluster.ID,
			Labels: map[string]string{
				"cluster":  cluster.Cluster.ID,
				"customer": cluster.Customer.ID,
			},
		},
	}

	if _, err := s.k8sClient.Core().Namespaces().Create(&namespace); err != nil && !apierrors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}
	return nil
}

func (s *Service) deleteClusterNamespace(cluster clustertpr.Spec) error {
	return s.k8sClient.Core().Namespaces().Delete(cluster.Cluster.ID, apismetav1.NewDeleteOptions(0))
}
