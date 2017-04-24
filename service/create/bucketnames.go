package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func (s *Service) bucketName(cluster awstpr.CustomObject) string {
	accountID := s.awsConfig.AccountID()
	region := cluster.Spec.AWS.Region
	customerID := cluster.Spec.Cluster.Customer.ID

	name := fmt.Sprintf("%s-g8s-%s-%s", accountID, region, customerID)

	return name
}

func (s *Service) bucketObjectDirPath(cluster awstpr.CustomObject) string {
	clusterID := cluster.Spec.Cluster.Cluster.ID
	return fmt.Sprintf("%s/cloudconfig", clusterID)
}

func (s *Service) bucketObjectFullDirPath(cluster awstpr.CustomObject) string {
	bucketName := s.bucketName(cluster)

	dirPath := s.bucketObjectDirPath(cluster)
	return fmt.Sprintf("%s/%s", bucketName, dirPath)
}

func (s *Service) bucketObjectName(cluster awstpr.CustomObject, prefix string) string {
	dirPath := s.bucketObjectDirPath(cluster)
	return fmt.Sprintf("%s/%s", dirPath, prefix)
}
