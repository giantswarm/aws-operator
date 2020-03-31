package tccp

// ConfigAPIWhitelist defines guest cluster k8s public/private api whitelisting.
type ConfigAPIWhitelist struct {
	Private ConfigAPIWhitelistSecurityGroup
	Public  ConfigAPIWhitelistSecurityGroup
}

// ConfigAPIWhitelistSecurityGroup represents the structure required for
// defining whitelisting for resource security group
type ConfigAPIWhitelistSecurityGroup struct {
	Enabled    bool
	SubnetList []string
}
