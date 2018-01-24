package legacyv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
)

func (s *Resource) createClusterNamespace(cluster v1alpha1.Cluster) error {
	namespace := v1.Namespace{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: cluster.ID,
			Labels: map[string]string{
				"cluster":  cluster.ID,
				"customer": cluster.Customer.ID,
			},
		},
	}

	if _, err := s.k8sClient.Core().Namespaces().Create(&namespace); err != nil && !apierrors.IsAlreadyExists(err) {
		return microerror.Mask(err)
	}
	return nil
}

func (s *Resource) deleteClusterNamespace(cluster v1alpha1.Cluster) error {
	return s.k8sClient.Core().Namespaces().Delete(cluster.ID, apismetav1.NewDeleteOptions(0))
}
