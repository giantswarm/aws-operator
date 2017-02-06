package operator

import (
	"github.com/giantswarm/clusterspec"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

func (s *Service) createClusterNamespace(cluster clusterspec.Cluster) error {
	namespace := v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: cluster.Name,
			Labels: map[string]string{
				"cluster":  cluster.Name,
				"customer": cluster.Spec.Customer,
			},
		},
	}

	if _, err := s.k8sclient.Core().Namespaces().Create(&namespace); err != nil && !errors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}
	return nil
}

func (s *Service) deleteClusterNamespace(cluster clusterspec.Cluster) error {
	return s.k8sclient.Core().Namespaces().Delete(cluster.Name, v1.NewDeleteOptions(0))
}
