package create

import (
	"context"
	"github.com/ericchiang/k8s/api/v1"
	"github.com/giantswarm/clustertpr"
	microerror "github.com/giantswarm/microkit/error"
)

func (s *Service) createClusterNamespace(cluster clustertpr.Cluster) error {
	namespace := &v1.Namespace{
		Metadata: &v1.ObjectMeta{
			Name: &cluster.Cluster.ID,
			Labels: map[string]string{
				"cluster":  cluster.Cluster.ID,
				"customer": cluster.Customer.ID,
			},
		},
	}

	if _, err := s.k8sClient.CoreV1().CreateNamespace(context.Background(), namespace); err != nil {
		return microerror.MaskAny(err)
	}
	return nil
}

func (s *Service) deleteClusterNamespace(cluster clustertpr.Cluster) error {
	return s.k8sClient.CoreV1().DeleteNamespace(context.Background(), cluster.Cluster.ID)
}
