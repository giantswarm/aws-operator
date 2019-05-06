package key

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func DockerVolumeResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("DockerVolume", cluster)
}

func MasterInstanceResourceName(cluster v1alpha1.Cluster) string {
	return getResourcenameWithTimeHash("MasterInstance", cluster)
}

// getResourcenameWithTimeHash returns a string cromprised of some prefix, a
// time hash and a cluster ID.
func getResourcenameWithTimeHash(prefix string, cluster v1alpha1.Cluster) string {
	id := strings.Replace(ClusterID(cluster), "-", "", -1)

	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	timeHash := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	upperTimeHash := strings.ToUpper(timeHash)
	upperClusterID := strings.ToUpper(id)

	return fmt.Sprintf("%s%s%s", prefix, upperClusterID, upperTimeHash)
}
