package label

const (
	// LegacyClusterID is an old style label for ClusterID
	LegacyClusterID = "clusterID"
	// LegacyClusterKey is an old style label to specify type of a secret that
	// is used for guest cluster. This is replaced by RandomKey.
	LegacyClusterKey = "clusterKey"
	// LegacyComponent is an old style label to identify which component a
	// specific CertConfig belongs to.
	LegacyComponent = "clusterComponent"
)
