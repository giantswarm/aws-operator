package randomkeytpr

// ClusterComponent represents the individual component of a k8s cluster, e.g.
// the API server, or etcd These are used when getting a secret from the k8s
// API, to identify the component the secret belongs to.
type ClusterComponent string

func (c ClusterComponent) String() string {
	return string(c)
}

// Key represents the type of Random Key asset, e.g. a encryption key.
// These are used when getting a secret from the k8s API, to
// identify the specific type of Random Key asset that is contained in the secret.
type Key string

func (c Key) String() string {
	return string(c)
}

// These constants are used to match each asset in the secret.
const (
	// EncryptionKey is the key for the kubernetes encryption.
	EncryptionKey Key = "encryption"
)

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// KeyLabel is the label used in the secret to identify a cluster
	// key.
	KeyLabel string = "clusterKey"
	// ClusterIDLabel is the label used in the secret to identify a cluster.
	ClusterIDLabel string = "clusterID"
)

// RandomKeyTypes is a slice enumerating all the Random Key assets we need to boot the
// cluster.
var RandomKeyTypes = []Key{
	EncryptionKey,
}

// ValidComponent looks for el among the components.
func ValidKey(el Key) bool {
	for _, v := range RandomKeyTypes {
		if el == v {
			return true
		}
	}
	return false
}

// CompactRandomKeyAssets is a struct used by operators to store stringified Random Key assets.
type CompactRandomKeyAssets struct {
	APIServerEncryptionKey string
}
