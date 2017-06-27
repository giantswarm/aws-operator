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

func (s *Service) bucketObjectURL(cluster awstpr.CustomObject, objectRelativePath string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName(cluster), objectRelativePath)
}
