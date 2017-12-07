package legacyv2

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (s *Resource) bucketName(cluster v1alpha1.AWSConfig) string {
	clusterID := keyv2.ClusterID(cluster)
	accountID := s.awsConfig.AccountID()

	name := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	return name
}

func (s *Resource) bucketObjectName(prefix string) string {
	return fmt.Sprintf("cloudconfig/%s", prefix)
}

func (s *Resource) bucketObjectURL(cluster v1alpha1.AWSConfig, objectRelativePath string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName(cluster), objectRelativePath)
}
