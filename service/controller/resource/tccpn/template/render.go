package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v13/pkg/template"
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
		TemplateMainRecordSets,
	}

	s, err := template.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
