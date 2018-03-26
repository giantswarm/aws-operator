package randomkeys

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// RandomKeyLabel is the label used in the secret to identify a secret
	// containing the random key
	//
	// TODO replace with "giantswarm.io/randomkey" and add to
	// https://github.com/giantswarm/fmt.
	RandomKeyLabel = "clusterKey"
	// ClusterIDLabel is the label used in the secret to identify a secret
	// containing the random key.
	//
	// TODO replace with "giantswarm.io/cluster-id"
	ClusterIDLabel = "clusterID"

	SecretNamespace = "default"
)

type key string

// These constants used as RandomKey
// parsing a secret received from the API.
const (
	EncryptionKey key = "encryption"
)

type RandomKey []byte

type Cluster struct {
	APIServerEncryptionKey RandomKey
}
