package kubeadmtokentpr

const (
	// ClusterIDLabel is a label of a secret in which we store cluster IDs.
	ClusterIDLabel = "clusterID"
	// KubeadmTokenKey is a key in a secret in which we store generated tokens.
	KubeadmTokenKey = "kubeadmToken"
)
