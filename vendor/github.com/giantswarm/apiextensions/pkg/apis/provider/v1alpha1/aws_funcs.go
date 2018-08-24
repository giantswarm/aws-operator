package v1alpha1

func (a AWSConfig) ClusterStatus() StatusCluster {
	return a.Status.Cluster
}
