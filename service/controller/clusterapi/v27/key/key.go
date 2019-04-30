package key

import (
	"fmt"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.%s", ClusterID(cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster v1alpha1.Cluster) string {
	return providerSpec(cluster).Cluster.DNS.Domain
}

func ClusterID(cluster v1alpha1.Cluster) string {
	return providerStatus(cluster).Cluster.ID
}

func ClusterNamespace(cluster v1alpha1.Cluster) string {
	return ClusterID(cluster)
}
