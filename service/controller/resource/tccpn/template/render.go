package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/template"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainAutoScalingGroup,
		TemplateMainENI,
		TemplateMainEtcdVolume,
		TemplateMainIAMPolicies,
		TemplateMainLaunchTemplate,
		TemplateMainOutputs,
	}

	s, err := template.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
