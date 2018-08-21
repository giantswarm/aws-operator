package render

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/microerror"
)

func Render(content string, data interface{}) (string, error) {
	t, err := template.New("e2etemplate").Parse(content)
	if err != nil {
		return "", microerror.Mask(err)
	}

	b := new(bytes.Buffer)
	err = t.Execute(b, data)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return b.String(), nil
}
