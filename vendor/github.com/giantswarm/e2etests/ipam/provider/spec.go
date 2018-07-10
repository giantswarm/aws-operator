package provider

type GuestClusterConfig struct {
	Name string
}

type Interface interface {
	// CreateCluster creates provider config chart deployment for guest cluster
	// which provider specific operator reconciles on and therefore yields new
	// guest cluster with given parameters. This function does not wait for
	// guest cluster to get ready. This only creates provider config for it.
	CreateCluster(clusterName string) error
	// DeleteCluster deletes provider config chart deployment for guest
	// cluster. Provider specific operator reconciles on it and deletes the
	// guest cluster eventually. This function does not wait for guest cluster
	// to get deleted but only deletes the chart deployment of provider config.
	DeleteCluster(clusterName string)
}
