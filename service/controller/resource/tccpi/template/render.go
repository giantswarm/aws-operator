package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v12/pkg/template"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainIAMRoles,
	}

	s, err := template.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
