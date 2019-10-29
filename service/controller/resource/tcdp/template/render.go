package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/internal/templates"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainIAMRoles,
	}

	s, err := templates.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
