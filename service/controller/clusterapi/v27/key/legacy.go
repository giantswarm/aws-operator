package key

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates/cloudformation/tccp"
)

// NOTE that code below is deprecated and needs refactoring.

func CloudConfigSmallTemplates() []string {
	return []string{
		cloudconfig.Small,
	}
}

func CloudFormationGuestTemplates() []string {
	return []string{
		tccp.AutoScalingGroup,
		tccp.IAMPolicies,
		tccp.Instance,
		tccp.InternetGateway,
		tccp.LaunchConfiguration,
		tccp.LoadBalancers,
		tccp.Main,
		tccp.NatGateway,
		tccp.LifecycleHooks,
		tccp.Outputs,
		tccp.RecordSets,
		tccp.RouteTables,
		tccp.SecurityGroups,
		tccp.Subnets,
		tccp.VPC,
	}
}
