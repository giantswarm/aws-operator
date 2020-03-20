package template

type ParamsMainSecurityGroups struct {
	APIInternalELBSecurityGroupName  string
	APIInternalELBSecurityGroupRules []SecurityGroupRule
	APIWhitelistEnabled              bool
	AWSCNISecurityGroupName          string
	PrivateAPIWhitelistEnabled       bool
	MasterSecurityGroupName          string
	MasterSecurityGroupRules         []SecurityGroupRule
	EtcdELBSecurityGroupName         string
	EtcdELBSecurityGroupRules        []SecurityGroupRule
}

type SecurityGroupRule struct {
	Description         string
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}
