package templates

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/microerror"
)

func Render(templates []string, data interface{}) (string, error) {
	var err error

	main := template.New("main")
	for _, t := range templates {
		main, err = main.Parse(t)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var b bytes.Buffer
	err = main.ExecuteTemplate(&b, "main", data)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return b.String(), nil
}
