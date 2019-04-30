package template

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainRecordSets,
		TemplateMainRouteTables,
	}

	s, err := templates.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
