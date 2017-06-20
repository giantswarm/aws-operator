package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func (s *Service) bucketName(cluster awstpr.CustomObject) string {
	accountID := s.awsConfig.AccountID()
	clusterID := cluster.Spec.Cluster.Cluster.ID
	region := cluster.Spec.AWS.Region

	name := fmt.Sprintf("%s-g8s-%s-%s", accountID, clusterID, region)

	return name
}

func (s *Service) bucketObjectDirPath(cluster awstpr.CustomObject) string {
	clusterID := cluster.Spec.Cluster.Cluster.ID
	return fmt.Sprintf("%s/cloudconfig", clusterID)
}

func (s *Service) bucketObjectName(prefix string) string {
	return fmt.Sprintf("%s", prefix)
}
