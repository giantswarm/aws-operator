package legacyv1

import (
	"fmt"

	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/awstpr"
)

func (s *Resource) bucketName(cluster awstpr.CustomObject) string {
	clusterID := keyv1.ClusterID(cluster)
	accountID := s.awsConfig.AccountID()

	name := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	return name
}

func (s *Resource) bucketObjectName(prefix string) string {
	return fmt.Sprintf("cloudconfig/%s", prefix)
}

func (s *Resource) bucketObjectURL(cluster awstpr.CustomObject, objectRelativePath string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName(cluster), objectRelativePath)
}
