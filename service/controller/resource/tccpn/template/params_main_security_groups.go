package template

type ParamsMainSecurityGroups struct {
	APIInternalELBSecurityGroupName  string
	APIInternalELBSecurityGroupRules []SecurityGroupRule
	APIWhitelistEnabled              bool
	AWSCNISecurityGroupID            string
	EtcdELBSecurityGroupName         string
	EtcdELBSecurityGroupRules        []SecurityGroupRule
	MasterSecurityGroupName          string
	MasterSecurityGroupRules         []SecurityGroupRule
	PrivateAPIWhitelistEnabled       bool
	VPCID                            string
}

type SecurityGroupRule struct {
	Description         string
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}
