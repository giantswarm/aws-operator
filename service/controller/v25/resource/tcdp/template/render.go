package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v25/templates"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainAutoScalingGroup,
		TemplateMainIAMPolicies,
		TemplateMainLaunchConfiguration,
		TemplateMainLifecycleHooks,
		TemplateMainOutputs,
		TemplateMainRouteTableAssociation,
		TemplateMainSecurityGroups,
		TemplateMainSubnets,
	}

	s, err := templates.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
