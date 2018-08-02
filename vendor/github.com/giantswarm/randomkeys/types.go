package randomkeys

import "fmt"

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// randomKeyLabel is the label used in the secret to identify a secret
	// containing the random key.
	randomKeyLabel = "giantswarm.io/randomkey"
	// clusterLabel is the label used in the secret to identify a secret
	// containing the random key.
	clusterLabel = "giantswarm.io/cluster"

	// legacyRandomKeyLabel is the label used in the secret to identify a secret
	// containing the random key.
	//
	// TODO replace with "giantswarm.io/randomkey".
	legacyRandomKeyLabel = "clusterKey"
	// legacyClusterIDLabel is the label used in the secret to identify a secret
	// containing the random key.
	//
	// TODO replace with "giantswarm.io/cluster-id".
	legacyClusterIDLabel = "clusterID"

	SecretNamespace = "default"
)

type Key string

func (k Key) String() string {
	return string(k)
}

const (
	EncryptionKey Key = "encryption"
)

var AllKeys = []Key{
	EncryptionKey,
}

type RandomKey []byte

type Cluster struct {
	APIServerEncryptionKey RandomKey
}

// K8sName returns Kubernetes object name for the certificate name and
// the key.
func K8sName(clusterID string, key Key) string {
	return fmt.Sprintf("%s-%s", clusterID, key)
}

// K8sLabels returns labels for the Kubernetes  object for the certificate name
// and the key.
func K8sLabels(clusterID string, key Key) map[string]string {
	return map[string]string{
		randomKeyLabel:       key.String(),
		clusterLabel:         clusterID,
		legacyRandomKeyLabel: key.String(),
		legacyClusterIDLabel: clusterID,
	}
}
