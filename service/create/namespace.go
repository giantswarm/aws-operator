package create

import (
	"github.com/giantswarm/clustertpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

func (s *Service) createClusterNamespace(cluster clustertpr.Cluster) error {
	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: cluster.Cluster.ID,
			Labels: map[string]string{
				"cluster":  cluster.Cluster.ID,
				"customer": cluster.Customer.ID,
			},
		},
	}

	if _, err := s.K8sClient.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}
	return nil
}

func (s *Service) deleteClusterNamespace(cluster clustertpr.Cluster) error {
	return s.K8sClient.Core().Namespaces().Delete(cluster.Cluster.ID, v1.NewDeleteOptions(0))
}
