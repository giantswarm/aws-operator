package template

type ParamsMainSecurityGroups struct {
	APIWhitelistEnabled        bool
	PrivateAPIWhitelistEnabled bool
	MasterSecurityGroupName    string
	MasterSecurityGroupRules   []SecurityGroupRule
	EtcdELBSecurityGroupName   string
	EtcdELBSecurityGroupRules  []SecurityGroupRule
	VPCID                      string
}

type SecurityGroupRule struct {
	Description         string
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}
