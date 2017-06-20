package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func (s *Service) bucketName(cluster awstpr.CustomObject) string {
	accountID := s.awsConfig.AccountID()
	clusterID := cluster.Spec.Cluster.Cluster.ID

	name := fmt.Sprintf("%s-g8s-%s-%s", accountID, clusterID)

	return name
}

func (s *Service) bucketObjectName(prefix string) string {
	return fmt.Sprintf("cloudconfig/%s", prefix)
}
