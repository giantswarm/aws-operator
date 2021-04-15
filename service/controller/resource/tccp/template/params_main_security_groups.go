package template

import "github.com/aws/aws-sdk-go/service/ec2"

type ParamsMainSecurityGroups struct {
	APIWhitelist                    ParamsMainSecurityGroupsAPIWhitelist
	ClusterID                       string
	ControlPlaneNATGatewayAddresses []*ec2.Address
	ControlPlaneVPCCIDR             string
	TenantClusterVPCCIDR            string
	TenantClusterCNICIDR            string
}

type ParamsMainSecurityGroupsAPIWhitelist struct {
	Private ParamsMainSecurityGroupsAPIWhitelistSecurityGroup
	Public  ParamsMainSecurityGroupsAPIWhitelistSecurityGroup
}

type ParamsMainSecurityGroupsAPIWhitelistSecurityGroup struct {
	Enabled    bool
	SubnetList []string
}
