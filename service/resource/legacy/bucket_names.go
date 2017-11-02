package legacy

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

func (s *Service) bucketName(cluster awstpr.CustomObject) string {
	accountID := s.awsConfig.AccountID()
	clusterID := key.ClusterID(cluster)

	name := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	return name
}

func (s *Service) bucketObjectName(prefix string) string {
	return fmt.Sprintf("cloudconfig/%s", prefix)
}

func (s *Service) bucketObjectURL(cluster awstpr.CustomObject, objectRelativePath string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName(cluster), objectRelativePath)
}
