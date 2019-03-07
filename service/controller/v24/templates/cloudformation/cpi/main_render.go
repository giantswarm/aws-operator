package cpi

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/templates"
)

func Render(v interface{}) (string, error) {
	l := []string{
		Main,
		IAMRoles,
	}

	s, err := templates.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}
