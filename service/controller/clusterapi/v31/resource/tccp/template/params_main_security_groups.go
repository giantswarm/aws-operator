package template

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type ParamsSecurityGroups struct {
	APIInternalELBSecurityGroupName  string
	APIInternalELBSecurityGroupRules []securityGroupRule
	APIWhitelistEnabled              bool
	PrivateAPIWhitelistEnabled       bool
	MasterSecurityGroupName          string
	MasterSecurityGroupRules         []securityGroupRule
	IngressSecurityGroupName         string
	IngressSecurityGroupRules        []securityGroupRule
	EtcdELBSecurityGroupName         string
	EtcdELBSecurityGroupRules        []securityGroupRule
}

type securityGroupRule struct {
	Description         string
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}