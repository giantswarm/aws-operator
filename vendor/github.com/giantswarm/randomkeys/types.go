package randomkeys

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// componentLabel is the label used in the secret to identify a secret
	// containing the random key
	//
	// TODO replace with "giantswarm.io/randomkey" and add to
	// https://github.com/giantswarm/fmt.
	randomkeyLabel = "clusterKey"
	// clusterIDLabel is the label used in the secret to identify a secret
	// containing the random key.
	//
	// TODO replace with "giantswarm.io/cluster-id"
	clusterIDLabel = "clusterID"

	SecretNamespace = "default"
)

type key string

// These constants used as RandomKey
// parsing a secret received from the API.
const (
	encryptionKey key = "encryption"
)

type RandomKey []byte

type Cluster struct {
	APIServerEncryptionKey RandomKey
}
