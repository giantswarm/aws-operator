package create

import (
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

func (s *Service) createClusterNamespace(cluster awstpr.CustomObject) error {
	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: cluster.Name,
			Labels: map[string]string{
				"cluster":  cluster.Name,
				"customer": cluster.Spec.Customer.ID,
			},
		},
	}

	if _, err := s.k8sClient.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}
	return nil
}

func (s *Service) deleteClusterNamespace(cluster awstpr.CustomObject) error {
	return s.k8sClient.Core().Namespaces().Delete(cluster.Name, v1.NewDeleteOptions(0))
}
