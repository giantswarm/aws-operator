package tccpn

// APIWhitelist defines guest cluster k8s public/private api whitelisting.
type APIWhitelist struct {
	Private Whitelist
	Public  Whitelist
}

// Whitelist represents the structure required for defining whitelisting for
// resource security group
type Whitelist struct {
	Enabled    bool
	SubnetList string
}
