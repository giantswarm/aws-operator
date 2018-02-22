package certs

type Interface interface {
	SearchCluster(clusterID string) (Cluster, error)
	SearchDraining(clusterID string) (Draining, error)
	SearchMonitoring(clusterID string) (Monitoring, error)
}
