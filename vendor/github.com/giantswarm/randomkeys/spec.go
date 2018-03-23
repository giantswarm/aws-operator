package randomkeys

type Interface interface {
	SearchCluster(clusterID string) (Cluster, error)
}
