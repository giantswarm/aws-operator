package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v14/pkg/template"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainAutoScalingGroup,
		TemplateMainIAMPolicies,
		TemplateMainLaunchTemplate,
		TemplateMainOutputs,
		TemplateMainRouteTables,
		TemplateMainSecurityGroups,
		TemplateMainSubnets,
		TemplateMainVPC,
	}

	s, err := template.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
