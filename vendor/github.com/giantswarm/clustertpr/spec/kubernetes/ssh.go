package kubernetes

type SSH struct {
	// PublicKeys is a list of SSH public keys being added to each Kubernetes
	// node. It can contain admin specific public keys as well as customer
	// specific ones.
	PublicKeys []string `json:"publicKeys" yaml:"publicKeys"`
}
